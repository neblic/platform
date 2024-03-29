package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"

	"github.com/IBM/sarama"
	"github.com/a8m/envsubst"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/neblic/platform/cmd/kafka-sampler/filter"
	"github.com/neblic/platform/cmd/kafka-sampler/neblic"
	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler"
	"go.uber.org/zap"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func decodeToStruct(i, o interface{}) error {
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			func(
				f reflect.Type,
				t reflect.Type,
				data interface{},
			) (interface{}, error) {
				// Check that the data is string. Standard hook logic
				if f.Kind() != reflect.String {
					return data, nil
				}

				// Check that the target type is a filter.Predicate interface.
				predicateType := reflect.TypeOf((*filter.Predicate)(nil)).Elem()
				if !t.Implements(predicateType) {
					return data, nil
				}

				predicate, err := filter.NewRegex(data.(string))
				if err != nil {
					return nil, fmt.Errorf("error parsing the regex predicate %s: %v", data.(string), err)
				}

				return predicate, nil
			},
		),
		TagName:              "",
		IgnoreUntaggedFields: false,
		Metadata:             nil,
		Result:               o,
		WeaklyTypedInput:     true,
		MatchName: func(f string, s string) bool {
			return strings.EqualFold(strings.ToLower(f), strings.ToLower(s))
		},
	}

	d, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}

	return d.Decode(i)
}

func initConfig(path *string) *Config {
	k := koanf.New(".")

	if path != nil {
		// Load YAML config file
		yamlConfig, err := os.ReadFile(*path)
		if err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}

		// Expand end vars
		yamlConfigExp, err := envsubst.Bytes(yamlConfig)
		if err != nil {
			log.Fatalf("Error expanding env vars in config file: %v", err)
		}

		// Load file contents
		if err := k.Load(rawbytes.Provider(yamlConfigExp), yaml.Parser()); err != nil {
			log.Fatalf("Error loading config file: %v", err)
		}
	}

	// Load env vars
	k.Load(env.Provider("", ".", func(s string) string {
		c := cases.Title(language.English)
		parts := strings.Split(s, "_")

		var titleParts []string
		for _, part := range parts {
			titleParts = append(titleParts, c.String(part))
		}

		return strings.Join(titleParts, ".")
	}), nil)

	// Decode back into the config struct overwritting default values
	config := NewConfig()
	if err := decodeToStruct(k.Raw(), config); err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	// Finalize configuration
	if err := config.Finalize(); err != nil {
		log.Fatalf("Error finalizing config: %v", err)
	}

	// We need to recreate the metric registry, otherwise it is not properly initilized and segfaults
	config.Kafka.Sarama.MetricRegistry = sarama.NewConfig().MetricRegistry

	return config
}

func initNeblic(ctx context.Context, logger logging.Logger, config *neblic.Config) error {
	logger.Info("Initializing Neblic connection", "config", config)

	// Propagate options
	providerOpts := []sampler.ProviderOption{
		sampler.WithLogger(logger),
	}
	if config.TLS {
		providerOpts = append(providerOpts, sampler.WithTLS())
	}
	if config.Bearer != "" {
		providerOpts = append(providerOpts, sampler.WithBearerAuth(config.Bearer))
	}

	provider, err := sampler.NewProvider(ctx, config.Settings, providerOpts...)
	if err != nil {
		return fmt.Errorf("error initializing neblic provider: %w", err)
	}
	err = sampler.SetProvider(provider)
	if err != nil {
		return fmt.Errorf("error setting global sampler provider: %w", err)
	}

	return nil
}

func runKafkaSampler(ctx context.Context, logger logging.Logger, config *Config) error {
	logger.Info("Initializing Kafka connection", "endpoints", config.Kafka.Servers)

	stdLog, err := zap.NewStdLogAt(logger.With("source", "sarama").ZapLogger().WithOptions(zap.AddCallerSkip(0)), zap.DebugLevel)
	if err != nil {
		logger.Error("Error initializing sarama logger, program will continue without outputing sarama debug logs", "error", err)
	}
	sarama.Logger = stdLog

	// Run Kafka Sampler
	kafkaSampler, err := NewKafkaSampler(ctx, logger, config)
	if err != nil {
		return fmt.Errorf("error initializing kafka sampler: %w", err)
	}
	err = kafkaSampler.Run()
	if err != nil {
		return fmt.Errorf("error running kafka sampler: %w", err)
	}

	return nil
}

func main() {
	var configPath = flag.String("config", "/etc/neblic/kafka-sampler/config.yaml", "configuration file path")
	flag.Parse()

	config := initConfig(configPath)
	logger, err := logging.NewZapProd(config.Logging.Level)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.ZapLogger().Sync()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		cancel()
	}()

	if err := initNeblic(ctx, logger, &config.Neblic); err != nil {
		log.Fatalf("Error initializing Neblic: %s", err)
	}

	if err := runKafkaSampler(ctx, logger, config); err != nil {
		log.Fatalf("Error starting: %s", err)
	}
}

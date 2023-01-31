package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/mitchellh/mapstructure"
	"github.com/neblic/platform/cmd/kafka-sampler/filter"
	"github.com/neblic/platform/cmd/kafka-sampler/neblic"
	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler"
	"github.com/neblic/platform/sampler/global"
	"github.com/spf13/viper"
)

func initViper() *Config {
	viper.SetConfigName("config")                     // name of config file (without extension)
	viper.SetConfigType("yaml")                       // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/neblic/kafka-sampler/") // path to look for the config file in
	viper.AddConfigPath(".")

	// Configuration parameters read from a config file or an environment variable
	// could mismatch the expected type in the configuration struct. A set of decode
	// hook funcions are used to automatically convert between types:
	// - string to time.Duration (e.g. 15s, 1m, etc.)
	// - string to []string using "," to split
	// - string to filter.Predicate
	decodeHookFunc := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		func() mapstructure.DecodeHookFuncType {
			return func(
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

				var predicate filter.Predicate
				before, after, found := strings.Cut(data.(string), ":")
				if found && before == "regex" {
					// The processed string contains a regex and it follows the 'regex:<regex>' pattern.
					var err error
					predicate, err = filter.NewRegex(after)
					if err != nil {
						return nil, fmt.Errorf("error parsing the regex predicate %s: %v", after, err)
					}
				} else {
					// The processed string contains a simple string
					predicate = filter.NewString(data.(string))
				}

				return predicate, nil
			}
		},
	)
	viper.DecodeHook(decodeHookFunc)

	// Inject default values (that also enables the usage of env vars to override the value)
	viper.SetDefault("verbose", false)
	viper.SetDefault("kafka.servers", []string{"localhost:9092"})
	viper.SetDefault("kafka.consumergroup", "kafkasampler")
	viper.SetDefault("neblic.resourcename", "kafkasampler")
	viper.SetDefault("neblic.controlserveraddr", "localhost:8899")
	viper.SetDefault("neblic.dataserveraddr", "localhost:4317")
	viper.SetDefault("neblic.updatestatsperiod", "15s")
	viper.SetDefault("reconcileperiod", time.Minute)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("config: error reading config file: " + err.Error())
	}

	// Allow the overwrite of all the provided keys using environment variables
	for _, key := range viper.AllKeys() {
		envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		err := viper.BindEnv(key, envKey)
		if err != nil {
			log.Fatal("config: unable to bind env: " + err.Error())
		}
	}

	config := NewConfig()
	if err := viper.Unmarshal(config); err != nil {
		log.Fatal("config: unable to decode into struct: " + err.Error())
	}

	return config
}

func initNeblic(ctx context.Context, logger logging.Logger, config *neblic.Config) {

	// Propagate options
	options := []sampler.Option{sampler.WithLogger(logger)}
	if config.Bearer != "" {
		options = append(options, sampler.WithBearerAuth(config.Bearer))
	}
	if config.TLS {
		options = append(options, sampler.WithTLS())
	}
	if config.SamplerLimit != 0 {
		options = append(options, sampler.WithSamplingRateBurst(int64(config.SamplerLimit)), sampler.WithSamplingRateLimit(int64(config.SamplerLimit)))
	}
	if config.UpdateStatsPeriod != 0 {
		options = append(options, sampler.WithUpdateStatsPeriod(config.UpdateStatsPeriod))
	}

	provider, err := sampler.NewProvider(ctx, config.Settings, options...)
	if err != nil {
		log.Panicf("Error initializing the neblic provider: %v", err)
	}
	err = global.SetSamplerProvider(provider)
	if err != nil {
		log.Panicf("Error setting global sampler provider: %v", err)
	}

}

func runKafkaSampler(ctx context.Context, logger logging.Logger, config *Config) {
	log.Println("Starting a new Sarama consumer")
	if config.Verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	// Run Kafka Sampler
	kafkaSampler, err := NewKafkaSampler(ctx, logger, config)
	if err != nil {
		logger.Error(err.Error())
	}
	err = kafkaSampler.Run()
	if err != nil {
		logger.Error(err.Error())
	}
}

func main() {

	config := initViper()

	// Initialize logger
	logger, _ := logging.NewZapDev()

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		cancel()
	}()

	initNeblic(ctx, logger, &config.Neblic)

	runKafkaSampler(ctx, logger, config)
}

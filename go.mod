module github.com/neblic/platform

go 1.21

require (
	github.com/axiomhq/hyperloglog v0.0.0-20240124082744-24bca3a5b39b
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/golang/protobuf v1.5.3
	github.com/google/cel-go v0.19.0
	github.com/google/go-cmp v0.6.0
	github.com/google/renameio/v2 v2.0.0
	github.com/google/uuid v1.6.0
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c
	github.com/onsi/ginkgo/v2 v2.15.0
	github.com/onsi/gomega v1.31.1
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/collector/component v0.94.1
	go.opentelemetry.io/collector/config/configopaque v0.94.1
	go.opentelemetry.io/collector/config/configtls v0.94.1
	go.opentelemetry.io/collector/consumer v0.94.1
	go.opentelemetry.io/collector/exporter v0.94.1
	go.opentelemetry.io/collector/exporter/otlpexporter v0.94.1
	go.opentelemetry.io/collector/extension v0.94.1
	go.opentelemetry.io/collector/extension/auth v0.94.1
	go.opentelemetry.io/collector/pdata v1.1.0
	go.opentelemetry.io/collector/processor v0.94.1
	go.opentelemetry.io/collector/semconv v0.94.1
	go.opentelemetry.io/otel/metric v1.23.1
	go.opentelemetry.io/otel/trace v1.23.1
	go.uber.org/atomic v1.11.0
	go.uber.org/zap v1.26.0
	golang.org/x/exp v0.0.0-20240205201215-2c58cdc269a3
	golang.org/x/oauth2 v0.17.0
	golang.org/x/time v0.5.0
	google.golang.org/genproto/googleapis/api v0.0.0-20240205150955-31a09d347014
	google.golang.org/grpc v1.61.0
	google.golang.org/protobuf v1.32.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/compute v1.24.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.4-0.20230617002413-005d2dfb6b68 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-metro v0.0.0-20211217172704-adc40b04c140 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/pprof v0.0.0-20240207164012-fb44976bdcd5 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	go.opentelemetry.io/collector v0.94.1 // indirect
	go.opentelemetry.io/collector/config/configauth v0.94.1 // indirect
	go.opentelemetry.io/collector/config/configcompression v0.94.1 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.94.1 // indirect
	go.opentelemetry.io/collector/config/confignet v0.94.1 // indirect
	go.opentelemetry.io/collector/config/configretry v0.94.1 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.94.1 // indirect
	go.opentelemetry.io/collector/config/internal v0.94.1 // indirect
	go.opentelemetry.io/collector/confmap v0.94.1 // indirect
	go.opentelemetry.io/collector/featuregate v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.48.0 // indirect
	go.opentelemetry.io/otel v1.23.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.18.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240205150955-31a09d347014 // indirect
)

exclude github.com/knadh/koanf v1.5.0

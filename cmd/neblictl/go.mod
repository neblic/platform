module github.com/neblic/platform/cmd/neblictl

go 1.21

toolchain go1.21.0

replace github.com/neblic/platform => ../../

require (
	github.com/c-bata/go-prompt v0.2.5
	github.com/google/uuid v1.6.0
	github.com/neblic/platform v0.0.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/term v1.1.0
	golang.org/x/exp v0.0.0-20231127185646-65229373498e
	golang.org/x/sys v0.15.0
)

require (
	cloud.google.com/go/compute v1.23.3 // indirect
	cloud.google.com/go/compute/metadata v0.2.4-0.20230617002413-005d2dfb6b68 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mattn/go-tty v0.0.5 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/oauth2 v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231127180814-3a041ad873d4 // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

exclude github.com/c-bata/go-prompt v0.2.6

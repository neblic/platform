module github.com/neblic/platform/cmd/neblictl

go 1.21

toolchain go1.21.0

replace github.com/neblic/platform => ../../

require (
	github.com/c-bata/go-prompt v0.2.5
	github.com/google/uuid v1.3.1
	github.com/neblic/platform v0.0.1
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pkg/term v1.1.0
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63
	golang.org/x/sys v0.15.0
)

require (
	cloud.google.com/go/compute v1.23.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.4-0.20230617002413-005d2dfb6b68 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mattn/go-tty v0.0.5 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.25.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/oauth2 v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
	google.golang.org/grpc v1.57.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

exclude github.com/c-bata/go-prompt v0.2.6

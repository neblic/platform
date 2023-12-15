package docs_test

import (
	"testing"

	// --8<-- [start:ProviderInitImport]
	"context"

	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler"
	"github.com/neblic/platform/sampler/defs"
	"github.com/neblic/platform/sampler/global"
	// --8<-- [end:ProviderInitImport]
)

// --8<-- [start:ProviderInit]
func initProvider(t *testing.T) defs.Provider {
	// the `Settings` struct contains the required configuration settings
	settings := sampler.Settings{
		ResourceName:      "service-name",
		ControlServerAddr: "otelcol:8899",
		DataServerAddr:    "otelcol:4317",
	}

	// additional options are provided with the `Options Pattern`
	logger, _ := logging.NewZapDev()
	provider, _ := sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))

	// optional: It is recommended to register the `Provider` as global.
	// this will allow you to initialize a `Sampler` from anywhere in your code
	// without needing a reference to the `Provider`.
	err := global.SetSamplerProvider(provider)
	if err != nil {
		t.Error(err)
	}

	return provider
}

// --8<-- [end:ProviderInit]

func TestInitProvider(t *testing.T) {
	initProvider(t)
}

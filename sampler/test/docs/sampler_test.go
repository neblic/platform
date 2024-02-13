package docs_test

import (
	"context"
	"testing"

	// --8<-- [start:SamplerInitImport]

	"github.com/neblic/platform/sampler"
	"github.com/neblic/platform/sampler/sample"
	// --8<-- [end:SamplerInitImport]
)

var someSampler sampler.Sampler

// --8<-- [start:SamplerInit]
func initSampler() sampler.Sampler {
	// initialize the schema that the sampled `Data Samples` will
	// a `DynamicSchema` supports Json strings and Go structs
	schema := sample.NewDynamicSchema()

	// creates a Sampler using the global provider
	// the global provider can be set using the `sampler.SetProvider(...)` method
	someSampler, _ := sampler.New("sampler-name", schema)

	return someSampler
}

// --8<-- [end:SamplerInit]

// --8<-- [start:SampleData]
func sampleData(ctx context.Context) bool {
	var dataSample string

	// evaluate a `Data Sample`
	return someSampler.Sample(ctx, sample.JSONSample(dataSample, sample.WithKey("key")))
}

// --8<-- [end:SampleData]

func TestSampler(_ *testing.T) {
	someSampler = initSampler()
	sampleData(context.Background())
}

package docs_test

import (
	"context"
	"testing"

	// --8<-- [start:SamplerInitImport]
	"github.com/neblic/platform/sampler"
	"github.com/neblic/platform/sampler/defs"
	// --8<-- [end:SamplerInitImport]
)

var someSampler defs.Sampler

// --8<-- [start:SamplerInit]
func initSampler() defs.Sampler {
	// initialize the schema that the sampled `Data Samples` will
	// a `DynamicSchema` supports Json strings and Go structs
	schema := defs.NewDynamicSchema()

	// equivalent to calling `global.SamplerProvider().Sampler()`
	someSampler, _ := sampler.Sampler("sampler-name", schema)

	return someSampler
}

// --8<-- [end:SamplerInit]

// --8<-- [start:SampleData]
func sampleData(ctx context.Context) bool {
	var dataSample string

	// evaluate a `Data Sample`
	return someSampler.Sample(ctx, defs.JsonSample(dataSample, ""))
}

// --8<-- [end:SampleData]

func TestSampler(t *testing.T) {
	someSampler = initSampler()
	sampleData(context.Background())
}

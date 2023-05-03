package global

import (
	"context"
	"testing"

	"github.com/neblic/platform/sampler/defs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockSampler struct {
	name   string
	schema defs.Schema
}

func (p *mockSampler) SampleJSON(ctx context.Context, jsonSample string) (bool, error) {
	return true, nil
}
func (p *mockSampler) SampleNative(ctx context.Context, nativeSample any) (bool, error) {
	return true, nil
}
func (p *mockSampler) SampleProto(ctx context.Context, protoSample proto.Message) (bool, error) {
	return true, nil
}
func (p *mockSampler) Sample(ctx context.Context, sample defs.Sample) bool {
	return true
}
func (p *mockSampler) Close() error {
	return nil
}

type mockProvider struct {
}

func (p *mockProvider) Sampler(name string, schema defs.Schema) (defs.Sampler, error) {
	return &mockSampler{name, schema}, nil
}

func TestSetSamplerProvider(t *testing.T) {
	// without setting a provider, samplers should be placeholders
	pp, err := SamplerProvider().Sampler("placeHolderSampler", defs.NewDynamicSchema())
	require.NoError(t, err)
	assert.IsType(t, &samplerPlaceholder{}, pp)

	// by default, placeholder samplers return false
	match := pp.Sample(context.Background(), defs.JsonSample("", ""))
	assert.Equal(t, false, match)

	// set mock provider
	err = SetSamplerProvider(&mockProvider{})
	require.NoError(t, err)

	// after setting the mock provider, samplers should have been replaced by mock samplers and return true
	match = pp.Sample(context.Background(), defs.JsonSample("", ""))
	assert.Equal(t, true, match)

	// new samplers should be mocks since it is what the mock provider returns
	pp2, err := SamplerProvider().Sampler("mockSampler", defs.NewDynamicSchema())
	require.NoError(t, err)
	assert.IsType(t, &mockSampler{}, pp2)
}

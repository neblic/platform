package sampler

import (
	"context"
	"testing"

	"github.com/neblic/platform/sampler/sample"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockSampler struct {
	name   string
	schema sample.Schema
}

func (p *mockSampler) SampleJSON(_ context.Context, _ string) (bool, error) {
	return true, nil
}
func (p *mockSampler) SampleNative(_ context.Context, _ any) (bool, error) {
	return true, nil
}
func (p *mockSampler) SampleProto(_ context.Context, _ proto.Message) (bool, error) {
	return true, nil
}
func (p *mockSampler) Sample(_ context.Context, _ sample.Sample) bool {
	return true
}
func (p *mockSampler) Close() error {
	return nil
}

type mockProvider struct {
}

func (p *mockProvider) Sampler(name string, schema sample.Schema, _ ...Option) (Sampler, error) {
	return &mockSampler{name, schema}, nil
}

func TestSetSamplerProvider(t *testing.T) {
	// without setting a provider, samplers should be placeholders
	pp, err := globalProvider().Sampler("placeHolderSampler", sample.NewDynamicSchema())
	require.NoError(t, err)
	assert.IsType(t, &samplerPlaceholder{}, pp)

	// by default, placeholder samplers return false
	match := pp.Sample(context.Background(), sample.JSONSample(""))
	assert.Equal(t, false, match)

	// set mock provider
	err = SetProvider(&mockProvider{})
	require.NoError(t, err)

	// after setting the mock provider, samplers should have been replaced by mock samplers and return true
	match = pp.Sample(context.Background(), sample.JSONSample(""))
	assert.Equal(t, true, match)

	// new samplers should be mocks since it is what the mock provider returns
	pp2, err := globalProvider().Sampler("mockSampler", sample.NewDynamicSchema())
	require.NoError(t, err)
	assert.IsType(t, &mockSampler{}, pp2)
}

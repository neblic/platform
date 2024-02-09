package sampler

import (
	"testing"

	"github.com/neblic/platform/controlplane/control"
	"github.com/stretchr/testify/assert"
)

func TestWithInitalValueDigest(t *testing.T) {
	// Create a new options object with empty initial configuration
	opts := &options{
		initialConfig: control.SamplerConfigUpdate{
			DigestUpdates: []control.DigestUpdate{},
			StreamUpdates: []control.StreamUpdate{},
		},
	}

	// Apply the WithInitalValueDigest option
	opt := WithInitalValueDigest(control.ComputationLocationCollector)
	opt.apply(opts)

	// Check that a default value digest and a default stream update were added
	assert.Equal(t, 1, len(opts.initialConfig.DigestUpdates))
	assert.Equal(t, control.DigestTypeValue, opts.initialConfig.DigestUpdates[0].Digest.Type)
	assert.Equal(t, valueDigestName, opts.initialConfig.DigestUpdates[0].Digest.Name)

	assert.Equal(t, 1, len(opts.initialConfig.StreamUpdates))
	assert.Equal(t, allStreamName, opts.initialConfig.StreamUpdates[0].Stream.Name)
	assert.Equal(t, allStreamCelRule, opts.initialConfig.StreamUpdates[0].Stream.StreamRule.Expression)
	assert.True(t, opts.initialConfig.StreamUpdates[0].Stream.ExportRawSamples)

	// If applied again, make sure that the default struct digest and stream update are not added again
	assert.Equal(t, 1, len(opts.initialConfig.DigestUpdates))
	assert.Equal(t, control.DigestTypeValue, opts.initialConfig.DigestUpdates[0].Digest.Type)
	assert.Equal(t, valueDigestName, opts.initialConfig.DigestUpdates[0].Digest.Name)

	assert.Equal(t, 1, len(opts.initialConfig.StreamUpdates))
	assert.Equal(t, allStreamName, opts.initialConfig.StreamUpdates[0].Stream.Name)
	assert.Equal(t, allStreamCelRule, opts.initialConfig.StreamUpdates[0].Stream.StreamRule.Expression)
	assert.True(t, opts.initialConfig.StreamUpdates[0].Stream.ExportRawSamples)
}

func TestWithInitalStructDigest(t *testing.T) {
	// Create a new options object with empty initial configuration
	opts := &options{
		initialConfig: control.SamplerConfigUpdate{
			DigestUpdates: []control.DigestUpdate{},
			StreamUpdates: []control.StreamUpdate{},
		},
	}

	// Apply the TestWithInitalStructDigest option
	opt := WithInitialStructDigest(control.ComputationLocationCollector)
	opt.apply(opts)

	// Check that a default struct digest and a default stream update were added
	assert.Equal(t, 1, len(opts.initialConfig.DigestUpdates))
	assert.Equal(t, control.DigestTypeSt, opts.initialConfig.DigestUpdates[0].Digest.Type)
	assert.Equal(t, structDigestName, opts.initialConfig.DigestUpdates[0].Digest.Name)

	assert.Equal(t, 1, len(opts.initialConfig.StreamUpdates))
	assert.Equal(t, allStreamName, opts.initialConfig.StreamUpdates[0].Stream.Name)
	assert.Equal(t, allStreamCelRule, opts.initialConfig.StreamUpdates[0].Stream.StreamRule.Expression)
	assert.True(t, opts.initialConfig.StreamUpdates[0].Stream.ExportRawSamples)

	// If applied again, make sure that the default struct digest and stream update are not added again
	WithInitialStructDigest(control.ComputationLocationCollector).apply(opts)
	assert.Equal(t, 1, len(opts.initialConfig.DigestUpdates))
	assert.Equal(t, control.DigestTypeSt, opts.initialConfig.DigestUpdates[0].Digest.Type)
	assert.Equal(t, structDigestName, opts.initialConfig.DigestUpdates[0].Digest.Name)

	assert.Equal(t, 1, len(opts.initialConfig.StreamUpdates))
	assert.Equal(t, allStreamName, opts.initialConfig.StreamUpdates[0].Stream.Name)
	assert.Equal(t, allStreamCelRule, opts.initialConfig.StreamUpdates[0].Stream.StreamRule.Expression)
	assert.True(t, opts.initialConfig.StreamUpdates[0].Stream.ExportRawSamples)
}

func TestWitTagsDLQ(t *testing.T) {
	// Initialize options
	o := &options{
		initialConfig: control.SamplerConfigUpdate{
			EventUpdates: []control.EventUpdate{},
		},
	}

	// Call WithTags with DLQTag
	opt := WithTags(DLQTag)
	opt.apply(o)

	// Assert that DLQ event is added
	assert.Equal(t, 1, len(o.initialConfig.EventUpdates))
	assert.Equal(t, DLQEventName, o.initialConfig.EventUpdates[0].Event.Name)
}

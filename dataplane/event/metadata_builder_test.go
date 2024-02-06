package event

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/neblic/platform/internal/pkg/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetadataBuilder(t *testing.T) {
	_, err := NewMetadataBuilder(`{'sampleKey': sample.key}`)
	assert.NoError(t, err)

	_, err = NewMetadataBuilder("{'sampleKey': nonExistingArg.key}")
	assert.Error(t, err)
}

func TestBuild(t *testing.T) {
	m, err := NewMetadataBuilder("{'sampleField': sample.field, 'key': key}")
	require.NoError(t, err)
	d := data.NewSampleDataFromJSON(`{"field": "value", "sensitiveField": "sensitiveValue"}`)

	str, err := m.Build(context.Background(), d, "sampleKey")
	assert.NoError(t, err)
	diff := cmp.Diff(`{"key":"sampleKey","sampleField":"value"}`, str, cmp.Transformer("ParseJSON", func(in []byte) (out any) {
		if err := json.Unmarshal(in, &out); err != nil {
			return err
		}
		return out
	}))
	assert.Empty(t, diff)
}

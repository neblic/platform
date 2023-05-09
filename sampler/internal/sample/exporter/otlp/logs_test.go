package otlp

import (
	"testing"
	"time"

	"github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/sampler/internal/sample/exporter"
	"github.com/stretchr/testify/assert"
	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

// TODO: test samples originating from proto and native?

func TestFromSamplerSamples(t *testing.T) {
	ts := time.Now()

	testCases := []struct {
		desc           string
		samplerSamples []exporter.SamplerSamples
	}{
		{
			desc: "empty resource samples",
		},
		{
			desc: "complete resource sample",
			samplerSamples: []exporter.SamplerSamples{
				{
					ResourceName: "some_resource_name1",
					SamplerName:  "some_sampler_name11",
					Samples: []exporter.Sample{
						{
							Ts:       ts,
							Type:     exporter.RawSampleType,
							Streams:  []data.SamplerStreamUID{"some_sampling_rule_uid1111"},
							Encoding: exporter.JSONSampleEncoding,
							Data:     []byte(`{"id": "some_id111"}`),
						},
						{
							Ts:       ts,
							Type:     exporter.RawSampleType,
							Streams:  []data.SamplerStreamUID{"some_sampling_rule_uid1211"},
							Encoding: exporter.JSONSampleEncoding,
							Data:     []byte(`{"id": "some_id121"}`),
						},
						{
							Ts:   ts,
							Type: exporter.RawSampleType,
							Streams: []data.SamplerStreamUID{
								"some_sampling_rule_uid1221",
								"some_sampling_rule_uid1222",
							},
							Encoding: exporter.JSONSampleEncoding,
							Data:     []byte(`{"id": "some_id122"}`),
						},
					},
				},
				{
					ResourceName: "some_resource_name2",
					SamplerName:  "some_sampler_name2",
					Samples: []exporter.Sample{
						{
							Ts:       ts,
							Type:     exporter.RawSampleType,
							Streams:  []data.SamplerStreamUID{"some_sampling_rule_uid2"},
							Encoding: exporter.JSONSampleEncoding,
							Data:     []byte(`{"id": "some_id2"}`),
						},
					},
				},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotLogs := fromSamplerSamples(tC.samplerSamples)

			assert.Equal(t, len(tC.samplerSamples), gotLogs.ResourceLogs().Len())
			for i := 0; i < gotLogs.ResourceLogs().Len(); i++ {
				cResourceLog := gotLogs.ResourceLogs().At(i)
				cResourceSample := tC.samplerSamples[i]

				samplerName, ok := cResourceLog.Resource().Attributes().Get(rlSamplerNameKey)
				assert.True(t, ok)

				name, ok := cResourceLog.Resource().Attributes().Get(conventions.AttributeServiceName)
				assert.True(t, ok)

				assert.Equal(t, cResourceSample.ResourceName, name.AsString())
				assert.Equal(t, cResourceSample.SamplerName, samplerName.AsString())

				assert.Equal(t, 1, cResourceLog.ScopeLogs().Len())
				cScopeLog := cResourceLog.ScopeLogs().At(0)

				assert.Equal(t, len(cResourceSample.Samples), cScopeLog.LogRecords().Len())
				for k := 0; k < cScopeLog.LogRecords().Len(); k++ {
					cLogRecord := cScopeLog.LogRecords().At(k)
					cSample := cResourceSample.Samples[k]

					assert.Equal(t, cLogRecord.Timestamp().AsTime(), cSample.Ts.UTC())
					assert.Equal(t, cLogRecord.Body().AsRaw(), cSample.Data)

					lrSamplingRuleUIDs, ok := cLogRecord.Attributes().Get(lrSampleStreamsUIDsKey)
					assert.True(t, ok)
					assert.Equal(t, len(cSample.Streams), lrSamplingRuleUIDs.Slice().Len())

					var samplingRuleUIDs []string
					for _, stream := range cSample.Streams {
						samplingRuleUIDs = append(samplingRuleUIDs, string(stream))
					}

					assert.ElementsMatch(t, samplingRuleUIDs, lrSamplingRuleUIDs.Slice().AsRaw())

					sampleTypeValue, ok := cLogRecord.Attributes().Get(lrSampleTypeKey)
					assert.Equal(t, cSample.Type.String(), sampleTypeValue.AsString())

					sampleEncodingValue, ok := cLogRecord.Attributes().Get(lrSampleTypeKey)
					assert.Equal(t, cSample.Type.String(), sampleEncodingValue.AsString())
				}
			}
		})
	}
}

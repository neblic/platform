package otlp

import (
	"testing"
	"time"

	"github.com/neblic/platform/sampler/internal/sample"
	"github.com/stretchr/testify/assert"
	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

func TestFromResourceSamples(t *testing.T) {
	ts := time.Now()

	testCases := []struct {
		desc            string
		resourceSamples []sample.ResourceSamples
	}{
		{
			desc: "empty resource samples",
		},
		{
			desc: "complete resource sample",
			resourceSamples: []sample.ResourceSamples{
				{
					ResourceName: "some_resource_name1",
					SamplerName:  "some_sampler_name11",
					SamplersSamples: []sample.SamplerSamples{
						{
							Samples: []sample.Sample{
								{
									Ts:   ts,
									Data: map[string]any{"id": "some_id111"},
									Matches: []sample.Match{
										{
											StreamUID: "some_sampling_rule_uid1111",
										},
									},
								},
							},
						},
						{
							Samples: []sample.Sample{
								{
									Ts:   ts,
									Data: map[string]any{"id": "some_id121"},
									Matches: []sample.Match{
										{
											StreamUID: "some_sampling_rule_uid1211",
										},
									},
								},
								{
									Ts:   ts,
									Data: map[string]any{"id": "some_id122"},
									Matches: []sample.Match{
										{
											StreamUID: "some_sampling_rule_uid1221",
										},
										{
											StreamUID: "some_sampling_rule_uid1222",
										},
									},
								},
							},
						},
					},
				},
				{
					ResourceName: "some_resource_name2",
					SamplerName:  "some_sampler_name2",
					SamplersSamples: []sample.SamplerSamples{
						{
							Samples: []sample.Sample{
								{
									Ts:   ts,
									Data: map[string]any{"id": "some_id2"},
									Matches: []sample.Match{
										{
											StreamUID: "some_sampling_rule_uid2",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotLogs := fromResourceSamples(tC.resourceSamples)

			assert.Equal(t, len(tC.resourceSamples), gotLogs.ResourceLogs().Len())
			for i := 0; i < gotLogs.ResourceLogs().Len(); i++ {
				cResourceLog := gotLogs.ResourceLogs().At(i)
				cResourceSample := tC.resourceSamples[i]

				samplerName, ok := cResourceLog.Resource().Attributes().Get(rlSamplerNameKey)
				assert.True(t, ok)

				name, ok := cResourceLog.Resource().Attributes().Get(conventions.AttributeServiceName)
				assert.True(t, ok)

				assert.Equal(t, cResourceSample.ResourceName, name.AsString())
				assert.Equal(t, cResourceSample.SamplerName, samplerName.AsString())

				assert.Equal(t, len(cResourceSample.SamplersSamples), cResourceLog.ScopeLogs().Len())
				for j := 0; j < cResourceLog.ScopeLogs().Len(); j++ {
					cScopeLog := cResourceLog.ScopeLogs().At(j)
					cSamplerSample := cResourceSample.SamplersSamples[j]

					scope := cScopeLog.Scope()
					scope.Attributes()

					assert.Equal(t, len(cSamplerSample.Samples), cScopeLog.LogRecords().Len())
					for k := 0; k < cScopeLog.LogRecords().Len(); k++ {
						cLogRecord := cScopeLog.LogRecords().At(k)
						cSample := cSamplerSample.Samples[k]

						assert.Equal(t, cLogRecord.Timestamp().AsTime(), cSample.Ts.UTC())
						assert.Equal(t, cLogRecord.Body().AsRaw(), cSample.Data)

						lrSamplingRuleUIDs, ok := cLogRecord.Attributes().Get(lrSamplingRuleUIDsKey)
						assert.True(t, ok)
						assert.Equal(t, len(cSample.Matches), lrSamplingRuleUIDs.Slice().Len())

						var samplingRuleUIDs []string
						for _, match := range cSample.Matches {
							samplingRuleUIDs = append(samplingRuleUIDs, string(match.StreamUID))
						}

						assert.ElementsMatch(t, samplingRuleUIDs, lrSamplingRuleUIDs.Slice().AsRaw())
					}
				}
			}
		})
	}
}

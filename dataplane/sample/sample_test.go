package sample

import (
	"testing"
	"time"

	"github.com/neblic/platform/controlplane/data"
	"github.com/stretchr/testify/assert"
)

// TODO: test samples originating from proto and native?

func TestFromSamplerSamples(t *testing.T) {
	ts := time.Now().UTC()

	testCases := []struct {
		desc           string
		samplerSamples []SamplerSamples
	}{
		{
			desc: "empty resource samples",
		},
		{
			desc: "complete resource sample",
			samplerSamples: []SamplerSamples{
				{
					ResourceName: "some_resource_name1",
					SamplerName:  "some_sampler_name11",
					Samples: []Sample{
						{
							Ts:       ts,
							Type:     RawType,
							Streams:  []data.SamplerStreamUID{"some_sampling_rule_uid1111"},
							Encoding: JSONEncoding,
							Data:     []byte(`{\"id\": \"some_id111\"}`),
						},
						{
							Ts:       ts,
							Type:     RawType,
							Streams:  []data.SamplerStreamUID{"some_sampling_rule_uid1211"},
							Encoding: JSONEncoding,
							Data:     []byte(`{\"id\": \"some_id121\"}`),
						},
						{
							Ts:   ts,
							Type: RawType,
							Streams: []data.SamplerStreamUID{
								"some_sampling_rule_uid1221",
								"some_sampling_rule_uid1222",
							},
							Encoding: JSONEncoding,
							Data:     []byte(`{\"id\": \"some_id122\"}`),
						},
					},
				},
				{
					ResourceName: "some_resource_name2",
					SamplerName:  "some_sampler_name2",
					Samples: []Sample{
						{
							Ts:       ts,
							Type:     RawType,
							Streams:  []data.SamplerStreamUID{"some_sampling_rule_uid2"},
							Encoding: JSONEncoding,
							Data:     []byte(`{"id": "some_id2"}`),
						},
					},
				},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotLogs := SamplesToOTLPLogs(tC.samplerSamples)
			parsedSamples := OTLPLogsToSamples(gotLogs)
			assert.Equal(t, tC.samplerSamples, parsedSamples)
		})
	}
}

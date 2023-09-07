package sample

import (
	"testing"
	"time"

	"github.com/neblic/platform/controlplane/control"
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
							Type:     control.RawSampleType,
							Streams:  []control.SamplerStreamUID{"some_sampling_rule_uid1111"},
							Encoding: JSONEncoding,
							Key:      "some_key111",
							Data:     []byte(`{\"id\": \"some_id111\"}`),
							Metadata: map[MetadataKey]string{
								EventUID:  "some_event_uid111",
								EventRule: "some_event_rule111",
							},
						},
						{
							Ts:       ts,
							Type:     control.RawSampleType,
							Streams:  []control.SamplerStreamUID{"some_sampling_rule_uid1211"},
							Encoding: JSONEncoding,
							Key:      "some_key121",
							Data:     []byte(`{\"id\": \"some_id121\"}`),
							Metadata: map[MetadataKey]string{
								EventUID:  "some_event_uid121",
								EventRule: "some_event_rule121",
							},
						},
						{
							Ts:   ts,
							Type: control.RawSampleType,
							Streams: []control.SamplerStreamUID{
								"some_sampling_rule_uid1221",
								"some_sampling_rule_uid1222",
							},
							Encoding: JSONEncoding,
							Key:      "some_key122",
							Data:     []byte(`{\"id\": \"some_id122\"}`),
							Metadata: map[MetadataKey]string{
								EventUID:  "some_event_uid122",
								EventRule: "some_event_rule122",
							},
						},
					},
				},
				{
					ResourceName: "some_resource_name2",
					SamplerName:  "some_sampler_name2",
					Samples: []Sample{
						{
							Ts:       ts,
							Type:     control.RawSampleType,
							Streams:  []control.SamplerStreamUID{"some_sampling_rule_uid2"},
							Encoding: JSONEncoding,
							Key:      "some_key2",
							Data:     []byte(`{"id": "some_id2"}`),
							Metadata: map[MetadataKey]string{
								EventUID:  "some_event_uid2",
								EventRule: "some_event_rule2",
							},
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

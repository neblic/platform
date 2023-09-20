package digest

import (
	"context"
	"testing"
	"time"

	"github.com/neblic/platform/controlplane/control"
	dpsample "github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testResourceName = "test_resource"
const testSamplerName = "test_sampler"

// TODO: use same timeout other tests
const testTimeout = 2 * time.Second

var testNotifyErr = func(t *testing.T) func(error) {
	return func(err error) {
		require.NoError(t, err)
	}
}

type mockExporter struct {
	exportedSamplerSamples []dpsample.SamplerSamples
}

func (e *mockExporter) Export(ctx context.Context, samplerSamples []dpsample.SamplerSamples) error {
	e.exportedSamplerSamples = append(e.exportedSamplerSamples, samplerSamples...)

	return nil
}

func (e *mockExporter) Close(context.Context) error {
	return nil
}

func TestBuildWorkers(t *testing.T) {
	tcs := map[string]struct {
		oldCfg, newCfg            map[control.SamplerDigestUID]control.Digest
		expectedWorkersDigestUIDs []control.SamplerDigestUID
	}{
		"New worker": {
			oldCfg: nil,
			newCfg: map[control.SamplerDigestUID]control.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: control.DigestTypeSt},
			},
			expectedWorkersDigestUIDs: []control.SamplerDigestUID{"sampler_digest_uid"},
		},
		"Delete worker": {
			oldCfg: map[control.SamplerDigestUID]control.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: control.DigestTypeSt},
			},
			newCfg:                    nil,
			expectedWorkersDigestUIDs: []control.SamplerDigestUID{},
		},
		// TODO: actually check that the worker settings have been updated
		"Update worker": {
			oldCfg: map[control.SamplerDigestUID]control.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: control.DigestTypeSt, St: control.DigestSt{MaxProcessedFields: 10}},
			},
			newCfg: map[control.SamplerDigestUID]control.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: control.DigestTypeSt, St: control.DigestSt{MaxProcessedFields: 20}},
			},
			expectedWorkersDigestUIDs: []control.SamplerDigestUID{"sampler_digest_uid"},
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			d := NewDigester(Settings{
				ResourceName:   testResourceName,
				SamplerName:    testSamplerName,
				EnabledDigests: []control.DigestType{control.DigestTypeSt},

				NotifyErr: testNotifyErr(t),
				Exporter:  &mockExporter{},
			})

			d.SetDigestsConfig(tc.oldCfg)
			d.SetDigestsConfig(tc.newCfg)

			var gotWorkersDigestUIDs []control.SamplerDigestUID
			for workerDigestUID := range d.workers {
				gotWorkersDigestUIDs = append(gotWorkersDigestUIDs, workerDigestUID)
			}

			assert.ElementsMatch(t, tc.expectedWorkersDigestUIDs, gotWorkersDigestUIDs)
		})
	}
}

func TestWorkerRun(t *testing.T) {
	tcs := map[string]struct {
		settings       workerSettings
		samples        []*data.Data
		expectedDigest dpsample.SamplerSamples
	}{
		"periodic digest": {
			settings: workerSettings{
				streamUID:      "stream_uid",
				resourceName:   testResourceName,
				samplerName:    testSamplerName,
				digest:         NewStDigest(10, testNotifyErr(t)),
				flushPeriod:    200 * time.Millisecond,
				inChBufferSize: 10,
				notifyErr:      testNotifyErr(t),
			},
			samples: []*data.Data{
				data.NewSampleDataFromJSON(`{ "field_double": 1 }`),
				data.NewSampleDataFromJSON(`{ "field_string": "some_string" }`),
			},
			expectedDigest: dpsample.SamplerSamples{
				ResourceName: testResourceName,
				SamplerName:  testSamplerName,
				Samples: []dpsample.Sample{
					{
						Type:     control.StructDigestSampleType,
						Streams:  []control.SamplerStreamUID{"stream_uid"},
						Encoding: dpsample.JSONEncoding,
						Data:     []byte(`{"obj":{"count":"2","fields":{"field_double":{"number":{"floatNum":{"count":"1"}}},"field_string":{"string":{"count":"1"}}}}}`),
					},
				},
			},
		},
	}
	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			testExporter := &mockExporter{}
			tc.settings.exporter = testExporter
			worker := newWorker(tc.settings)
			go worker.run()
			for _, sample := range tc.samples {
				worker.processSample(sample)
			}

			require.Eventually(t,
				func() bool { return len(testExporter.exportedSamplerSamples) >= 1 },
				testTimeout, 50*time.Millisecond,
			)
			require.Len(t, testExporter.exportedSamplerSamples[0].Samples, 1)

			// do not check the sample timestamp
			testExporter.exportedSamplerSamples[0].Samples[0].Ts = time.Time{}

			// do not check the JSON body with the Equal assertion
			expectedJSONDigest := string(tc.expectedDigest.Samples[0].Data)
			tc.expectedDigest.Samples[0].Data = nil
			gotJSONDigest := string(testExporter.exportedSamplerSamples[0].Samples[0].Data)
			testExporter.exportedSamplerSamples[0].Samples[0].Data = nil

			assert.Equal(t, tc.expectedDigest, testExporter.exportedSamplerSamples[0])
			assert.JSONEq(t, expectedJSONDigest, gotJSONDigest)

			worker.stop()
		})
	}
}

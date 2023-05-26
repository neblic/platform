package digest

import (
	"context"
	"testing"
	"time"

	"github.com/neblic/platform/controlplane/data"
	"github.com/neblic/platform/sampler/internal/sample"
	"github.com/neblic/platform/sampler/internal/sample/exporter"
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
	exportedSamplerSamples []exporter.SamplerSamples
}

func (e *mockExporter) Export(ctx context.Context, samplerSamples []exporter.SamplerSamples) error {
	e.exportedSamplerSamples = append(e.exportedSamplerSamples, samplerSamples...)

	return nil
}

func (e *mockExporter) Close(context.Context) error {
	return nil
}

func TestBuildWorkers(t *testing.T) {
	tcs := map[string]struct {
		oldCfg, newCfg            map[data.SamplerDigestUID]data.Digest
		expectedWorkersDigestUIDs []data.SamplerDigestUID
	}{
		"New worker": {
			oldCfg: nil,
			newCfg: map[data.SamplerDigestUID]data.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: data.DigestTypeSt},
			},
			expectedWorkersDigestUIDs: []data.SamplerDigestUID{"sampler_digest_uid"},
		},
		"Delete worker": {
			oldCfg: map[data.SamplerDigestUID]data.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: data.DigestTypeSt},
			},
			newCfg:                    nil,
			expectedWorkersDigestUIDs: []data.SamplerDigestUID{},
		},
		// TODO: actually check that the worker settings have been updated
		"Update worker": {
			oldCfg: map[data.SamplerDigestUID]data.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: data.DigestTypeSt, St: data.DigestSt{MaxProcessedFields: 10}},
			},
			newCfg: map[data.SamplerDigestUID]data.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: data.DigestTypeSt, St: data.DigestSt{MaxProcessedFields: 20}},
			},
			expectedWorkersDigestUIDs: []data.SamplerDigestUID{"sampler_digest_uid"},
		},
	}

	for name, tc := range tcs {
		tc := tc
		t.Run(name, func(t *testing.T) {
			d := NewDigester(Settings{
				ResourceName: testResourceName,
				SamplerName:  testSamplerName,

				NotifyErr: testNotifyErr(t),
				Exporter:  &mockExporter{},
			})

			d.SetDigestsConfig(tc.oldCfg)
			d.SetDigestsConfig(tc.newCfg)

			var gotWorkersDigestUIDs []data.SamplerDigestUID
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
		samples        []*sample.Data
		expectedDigest exporter.SamplerSamples
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
			samples: []*sample.Data{
				sample.NewSampleDataFromJSON(`{ "field_double": 1 }`),
				sample.NewSampleDataFromJSON(`{ "field_string": "some_string" }`),
			},
			expectedDigest: exporter.SamplerSamples{
				ResourceName: testResourceName,
				SamplerName:  testSamplerName,
				Samples: []exporter.Sample{
					{
						Type:     exporter.StructDigestSampleType,
						Streams:  []data.SamplerStreamUID{"stream_uid"},
						Encoding: exporter.JSONSampleEncoding,
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

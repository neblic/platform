package digest

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/neblic/platform/controlplane/control"
	dpsample "github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/internal/pkg/data"
	"github.com/neblic/platform/logging"
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
	exportedOtlpLogs []dpsample.OTLPLogs
}

func (e *mockExporter) Export(_ context.Context, otlpLogs dpsample.OTLPLogs) error {
	e.exportedOtlpLogs = append(e.exportedOtlpLogs, otlpLogs)

	return nil
}

func (e *mockExporter) Close(_ context.Context) error {
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
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: control.DigestTypeSt, St: &control.DigestSt{MaxProcessedFields: 100}},
			},
			expectedWorkersDigestUIDs: []control.SamplerDigestUID{"sampler_digest_uid"},
		},
		"Delete worker": {
			oldCfg: map[control.SamplerDigestUID]control.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: control.DigestTypeSt, St: &control.DigestSt{MaxProcessedFields: 100}},
			},
			newCfg:                    nil,
			expectedWorkersDigestUIDs: []control.SamplerDigestUID{},
		},
		// TODO: actually check that the worker settings have been updated
		"Update worker": {
			oldCfg: map[control.SamplerDigestUID]control.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: control.DigestTypeSt, St: &control.DigestSt{MaxProcessedFields: 10}},
			},
			newCfg: map[control.SamplerDigestUID]control.Digest{
				"sampler_digest_uid": {UID: "sampler_digest_uid", Type: control.DigestTypeSt, St: &control.DigestSt{MaxProcessedFields: 20}},
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
				Logger:    logging.NewNopLogger(),
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
		settings workerSettings
		samples  []*data.Data
		wantFn   func() dpsample.OTLPLogs
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
			wantFn: func() dpsample.OTLPLogs {
				otlpLogs := dpsample.NewOTLPLogs()
				samplerOtlpLogs := otlpLogs.AppendSamplerOTLPLogs(testResourceName, testSamplerName)
				structDigest := samplerOtlpLogs.AppendStructDigestOTLPLog()
				structDigest.SetStreamUIDs([]control.SamplerStreamUID{"stream_uid"})
				structDigest.SetSampleRawData(dpsample.JSONEncoding, []byte(`{"obj":{"count":"2","fields":{"field_double":{"number":{"floatNum":{"count":"1"}}},"field_string":{"string":{"count":"1"}}}}}`))

				return otlpLogs
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
				func() bool { return len(testExporter.exportedOtlpLogs) >= 1 },
				testTimeout, 50*time.Millisecond,
			)
			assert.Equal(t, testExporter.exportedOtlpLogs[0].Len(), 1)

			var wantResource string
			var wantSampler string
			var want dpsample.StructDigestOTLPLog
			dpsample.RangeWithType[dpsample.StructDigestOTLPLog](tc.wantFn(), func(resource, sample string, structDigestOtlp dpsample.StructDigestOTLPLog) {
				wantResource = resource
				wantSampler = sample
				want = structDigestOtlp
			})

			var gotResource string
			var gotSampler string
			var got dpsample.StructDigestOTLPLog
			dpsample.RangeWithType[dpsample.StructDigestOTLPLog](testExporter.exportedOtlpLogs[0], func(resource, sampler string, structDigestOtlp dpsample.StructDigestOTLPLog) {
				gotResource = resource
				gotSampler = sampler
				got = structDigestOtlp
			})

			assert.Equal(t, wantResource, gotResource)
			assert.Equal(t, wantSampler, gotSampler)
			assert.Equal(t, want.StreamUIDs(), got.StreamUIDs())
			diff := cmp.Diff(want.SampleRawData(), got.SampleRawData(), cmp.Transformer("ParseJSON", func(in []byte) (out any) {
				if err := json.Unmarshal(in, &out); err != nil {
					return err
				}
				return out
			}))
			assert.Empty(t, diff)

			worker.stop()
		})
	}
}

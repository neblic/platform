package sampler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"testing"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/dataplane/protos/test"
	dpsample "github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler/sample"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
	"google.golang.org/protobuf/proto"
)

/* Benchamrk description:
// --8<-- [start:BenchmarkDescription]
This benchmark suite measures the performance of the `Sampler.Sample()` method when sampling `Data Samples` with different number of fields, sizes and types.

Each benchmark is composed of 8 cases. The first 4 cases sample `Data Samples` with 4, 8, 16, and 32 fields. The last 4 cases have the same number of fields as the first 4, but with random byte filling to grow the `Data Sample` total size. The idea is to show that both variables, the number of fields and the total sample size, have an impact on the sampling performance.

The `BenchmarkStdLogger` is shown as a reference to compare the sampling performance with the standard library logger logging a message. And all tests write the sampled data to dev/null to avoid any I/O overhead.
// --8<-- [end:BenchmarkDescription]
*/

// each nested sample has 4 fields
func generateSample(numNestedSamples, byteFilling int) *test.TestSample {
	if numNestedSamples < 4 {
		numNestedSamples = 4
	}

	baseSamples := make([]*test.TestSample, numNestedSamples)
	for i := range baseSamples {
		bytes := make([]byte, byteFilling)
		for i := range bytes {
			bytes[i] = byte(rand.Intn(94) + 32)
		}

		baseSamples[i] = &test.TestSample{
			Double:  1.0,
			Bool:    true,
			String_: "shortstring",
			Bytes:   bytes,
		}
	}

	return &test.TestSample{
		Int32:      1,
		NestedMsgs: baseSamples,
	}
}

func generateJSONSamples() ([]int, []string) {
	numFields := make([]int, 8)
	samples := make([]string, 8)

	byteFilling := 0
	numNestedSamples := 8
	for i := range samples {
		// last 4 samples have 1KB of byte filling
		if i == 4 {
			byteFilling = 1024
			numNestedSamples = 8
		}

		jsonObj := generateSample(numNestedSamples, byteFilling)
		jsonStr, err := json.Marshal(jsonObj)
		if err != nil {
			panic(err)
		}

		numFields[i] = numNestedSamples * 4
		samples[i] = string(jsonStr)
		numNestedSamples *= 2
	}

	return numFields, samples
}

func BenchmarkStdLogger(b *testing.B) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	defer devNull.Close()

	writer := io.Writer(devNull)
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{})
	logger := slog.New(handler)

	numFields, jsonSamples := generateJSONSamples()
	for i, jsonSample := range jsonSamples {
		b.Run(fmt.Sprintf("%d_%dB", numFields[i], len(jsonSample)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				logger.Info(jsonSample)
			}
		})
	}
}

type mockExporter struct {
	writer io.Writer
}

func newMockExporter() *mockExporter {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	return &mockExporter{
		writer: devNull,
	}
}

func (e *mockExporter) Export(_ context.Context, otlpLogs dpsample.OTLPLogs) error {
	marshaler := plog.ProtoMarshaler{}
	logs, err := marshaler.MarshalLogs(otlpLogs.Logs())
	if err != nil {
		return err
	}

	_, err = e.writer.Write(logs)
	if err != nil {
		return err
	}

	return nil
}

func (e *mockExporter) Close(_ context.Context) error {
	return nil
}

type testCase struct {
	name   string
	config control.SamplerConfig
}

func testList(matchRule string) []testCase {
	return []testCase{
		{
			name:   "disabled",
			config: control.SamplerConfig{},
		},
		{
			name: "stream_all",
			config: control.SamplerConfig{
				Streams: control.Streams{
					"1": control.Stream{
						UID:  "1",
						Name: "all",
						StreamRule: control.Rule{
							Lang:       control.SrlCel,
							Expression: "true",
						},
					},
				},
			},
		},
		{
			name: "stream_expr_match",
			config: control.SamplerConfig{
				Streams: control.Streams{
					"1": control.Stream{
						UID:  "1",
						Name: "all",
						StreamRule: control.Rule{
							Lang:       control.SrlCel,
							Expression: matchRule,
						},
					},
				},
			},
		},
		{
			name: "stream_all_structure_digest",
			config: control.SamplerConfig{
				Streams: control.Streams{
					"1": control.Stream{
						UID:  "1",
						Name: "all",
						StreamRule: control.Rule{
							Lang:       control.SrlCel,
							Expression: "true",
						},
					},
				},
				Digests: control.Digests{
					"1": control.Digest{
						StreamUID:           "1",
						UID:                 "1",
						Name:                "st",
						Type:                control.DigestTypeSt,
						ComputationLocation: control.ComputationLocationSampler,
						St:                  &control.DigestSt{},
					},
				},
			},
		},
		{
			name: "stream_all_value_digest",
			config: control.SamplerConfig{
				Streams: control.Streams{
					"1": control.Stream{
						UID:  "1",
						Name: "all",
						StreamRule: control.Rule{
							Lang:       control.SrlCel,
							Expression: "true",
						},
					},
				},
				Digests: control.Digests{
					"1": control.Digest{
						StreamUID:           "1",
						UID:                 "1",
						Name:                "st",
						Type:                control.DigestTypeValue,
						ComputationLocation: control.ComputationLocationSampler,
						Value:               &control.DigestValue{},
					},
				},
			},
		},
		{
			name: "stream_all_export_raw",
			config: control.SamplerConfig{
				Streams: control.Streams{
					"1": control.Stream{
						UID:  "1",
						Name: "all",
						StreamRule: control.Rule{
							Lang:       control.SrlCel,
							Expression: "true",
						},
						ExportRawSamples: true,
					},
				},
			},
		},
		{
			name: "stream_expr_match_export_raw",
			config: control.SamplerConfig{
				Streams: control.Streams{
					"1": control.Stream{
						UID:  "1",
						Name: "all",
						StreamRule: control.Rule{
							Lang:       control.SrlCel,
							Expression: matchRule,
						},
						ExportRawSamples: true,
					},
				},
			},
		},
	}
}

func BenchmarkJSONSample(b *testing.B) {
	numFields, samples := generateJSONSamples()

	for _, tt := range testList("sample.int32==1") {
		for i, jsonSample := range samples {
			logger := logging.NewNopLogger()
			s, err := New(
				&Settings{
					Schema:           sample.NewDynamicSchema(),
					ControlPlaneAddr: "localhost:8899",
					LogsExporter:     newMockExporter(),
				},
				logger,
			)
			require.NoError(b, err)

			s.updateConfig(tt.config)
			s.digester.SetSync(true)

			ctx := context.Background()
			b.Run(fmt.Sprintf("%s/%d_%dB", tt.name, numFields[i], len(jsonSample)), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					smpl := sample.JSONSample(jsonSample)
					s.Sample(ctx, smpl)
				}
			})
		}
	}
}

func generateProtoSamples() ([]int, []proto.Message) {
	numFields := make([]int, 8)
	samples := make([]proto.Message, 8)

	byteFilling := 0
	numNestedSamples := 8
	for i := range samples {
		// last 4 samples have 1KB of byte filling
		if i == 4 {
			byteFilling = 1024
			numNestedSamples = 8
		}

		obj := generateSample(numNestedSamples, byteFilling)
		numFields[i] = numNestedSamples * 4
		samples[i] = obj

		numNestedSamples *= 2
	}

	return numFields, samples
}

func BenchmarkProtoSample(b *testing.B) {
	numFields, samples := generateProtoSamples()

	for _, tt := range testList("sample.int32==1") {
		for i, protoSample := range samples {
			logger := logging.NewNopLogger()
			s, err := New(
				&Settings{
					Schema:           sample.NewProtoSchema(&test.TestSample{}),
					ControlPlaneAddr: "localhost:8899",
					LogsExporter:     newMockExporter(),
				},
				logger,
			)
			require.NoError(b, err)

			s.updateConfig(tt.config)
			s.digester.SetSync(true)

			jsonSample, err := json.Marshal(protoSample)
			require.NoError(b, err)

			ctx := context.Background()
			b.Run(fmt.Sprintf("%s/%d_%dB", tt.name, numFields[i], len(jsonSample)), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					smpl := sample.ProtoSample(protoSample)
					s.Sample(ctx, smpl)
				}
			})
		}
	}
}

func BenchmarkNativeSample(b *testing.B) {
	// for now, we can use the proto struct as the native sample
	numFields, samples := generateProtoSamples()

	for _, tt := range testList("sample.Int32==1") {
		for i, protoSample := range samples {
			logger := logging.NewNopLogger()
			s, err := New(
				&Settings{
					Schema:           sample.NewDynamicSchema(),
					ControlPlaneAddr: "localhost:8899",
					LogsExporter:     newMockExporter(),
				},
				logger,
			)
			require.NoError(b, err)

			s.updateConfig(tt.config)
			s.digester.SetSync(true)

			jsonSample, err := json.Marshal(protoSample)
			require.NoError(b, err)

			ctx := context.Background()
			b.Run(fmt.Sprintf("%s/%d_%dB", tt.name, numFields[i], len(jsonSample)), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					smpl := sample.NativeSample(protoSample)
					s.Sample(ctx, smpl)
				}
			})
		}
	}
}

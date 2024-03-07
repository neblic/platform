package test_test

//revive:disable:dot-imports
import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/dataplane"
	"github.com/neblic/platform/dataplane/mock"
	"github.com/neblic/platform/dataplane/sample"
	"github.com/neblic/platform/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
)

func TestDataplane(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Sampler")
}

var _ = Describe("DataPlane", func() {
	var (
		logger logging.Logger

		eventExporter  *mock.LogsExporter
		sampleExporter *mock.LogsExporter
		digestExporter *mock.LogsExporter
		metricExporter *mock.MetricsExporter
		processor      *dataplane.Processor
	)

	BeforeEach(func() {
		var err error

		logger, err = logging.NewZapDev()
		Expect(err).ToNot(HaveOccurred())

		eventExporter = mock.NewLogsExporter()
		sampleExporter = mock.NewLogsExporter()
		digestExporter = mock.NewLogsExporter()
		metricExporter = mock.NewMetricsExporter()
		settings := &dataplane.Settings{
			Logger:         logger,
			ControlPlane:   nil,
			EventExporter:  eventExporter,
			SampleExporter: sampleExporter,
			DigestExporter: digestExporter,
			MetricExporter: metricExporter,
		}
		processor = dataplane.NewProcessor(settings)
	})

	Describe("Exporting digests", func() {

		When("there is a structure digest with sampler computation location", func() {
			It("should not generate structure digest samples", func() {
				streamUID := control.SamplerStreamUID(uuid.NewString())
				digestUID := control.SamplerDigestUID(uuid.NewString())
				processor.UpdateConfig("resource1", "sampler1", &control.SamplerConfig{
					Streams: map[control.SamplerStreamUID]control.Stream{
						streamUID: {
							UID:  streamUID,
							Name: "stream1",
							StreamRule: control.Rule{
								Lang: control.SrlCel, Expression: "sample.id==1",
							},
						},
					},
					Digests: map[control.SamplerDigestUID]control.Digest{
						digestUID: {
							UID:                 digestUID,
							StreamUID:           streamUID,
							FlushPeriod:         200 * time.Millisecond,
							ComputationLocation: control.ComputationLocationSampler,
							Type:                control.DigestTypeSt,
							St: &control.DigestSt{
								MaxProcessedFields: 100,
							},
						},
					},
				}, logger)

				// Process log
				logs := sample.NewOTLPLogs()
				samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")
				rawSampleLog := samplerLogs.AppendRawSampleOTLPLog()
				rawSampleLog.SetStreamUIDs([]control.SamplerStreamUID{streamUID})
				rawSampleLog.SetSampleRawData(sample.JSONEncoding, []byte(`{"id": 1}`))

				processor.ComputeDigests(logs)

				require.Never(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return len(digestExporter.StructDigests) == 1
					},
					time.Millisecond*500, time.Millisecond,
				)
			})
		})

		When("there is a structure digest with sampler computation location", func() {
			It("should export structure digest samples", func() {
				streamUID := control.SamplerStreamUID(uuid.NewString())
				digestUID := control.SamplerDigestUID(uuid.NewString())
				processor.UpdateConfig("resource1", "sampler1", &control.SamplerConfig{
					Streams: map[control.SamplerStreamUID]control.Stream{
						streamUID: {
							UID:  streamUID,
							Name: "stream1",
							StreamRule: control.Rule{
								Lang: control.SrlCel, Expression: "sample.id==1",
							},
						},
					},
					Digests: map[control.SamplerDigestUID]control.Digest{
						digestUID: {
							UID:                 digestUID,
							StreamUID:           streamUID,
							FlushPeriod:         200 * time.Millisecond,
							ComputationLocation: control.ComputationLocationCollector,
							Type:                control.DigestTypeSt,
							St: &control.DigestSt{
								MaxProcessedFields: 100,
							},
						},
					},
				}, logger)

				// Process log
				logs := sample.NewOTLPLogs()
				samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")
				rawSampleLog := samplerLogs.AppendRawSampleOTLPLog()
				rawSampleLog.SetStreamUIDs([]control.SamplerStreamUID{streamUID})
				rawSampleLog.SetSampleRawData(sample.JSONEncoding, []byte(`{"id": 1}`))

				processor.ComputeDigests(logs)

				// wait until the receiver has received the digest
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return len(digestExporter.StructDigests) == 1
					},
					time.Second, time.Millisecond)
			})
		})
	})
})

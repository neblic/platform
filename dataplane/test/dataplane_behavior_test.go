package test_test

//revive:disable:dot-imports
import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/dataplane"
	"github.com/neblic/platform/dataplane/mock"
	"github.com/neblic/platform/dataplane/sample"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDataplane(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Sampler")
}

var _ = Describe("DataPlane", func() {
	var (
		logger *zap.Logger

		exporter  *mock.Exporter
		processor *dataplane.Processor
	)

	BeforeEach(func() {
		var err error

		logger, err = zap.NewDevelopment()
		Expect(err).ToNot(HaveOccurred())

		exporter = mock.NewExporter()
		processor = dataplane.NewProcessor(logger, nil, exporter)
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
							St: control.DigestSt{
								MaxProcessedFields: 100,
							},
						},
					},
				})

				// Process log
				logs := sample.NewOTLPLogs()
				samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")
				rawSampleLog := samplerLogs.AppendRawSampleOTLPLog()
				rawSampleLog.SetStreams([]control.SamplerStreamUID{streamUID})
				rawSampleLog.SetSampleRawData(sample.JSONEncoding, []byte(`{"id": 1}`))
				err := processor.Process(context.Background(), logs)
				Expect(err).ToNot(HaveOccurred())

				require.Never(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return len(exporter.StructDigests) == 1
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
							St: control.DigestSt{
								MaxProcessedFields: 100,
							},
						},
					},
				})

				// Process log
				logs := sample.NewOTLPLogs()
				samplerLogs := logs.AppendSamplerOTLPLogs("resource1", "sampler1")
				rawSampleLog := samplerLogs.AppendRawSampleOTLPLog()
				rawSampleLog.SetStreams([]control.SamplerStreamUID{streamUID})
				rawSampleLog.SetSampleRawData(sample.JSONEncoding, []byte(`{"id": 1}`))

				err := processor.Process(context.Background(), logs)
				Expect(err).ToNot(HaveOccurred())

				// wait until the receiver has received the digest
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return len(exporter.StructDigests) == 1
					},
					time.Second, time.Millisecond)
			})
		})
	})
})

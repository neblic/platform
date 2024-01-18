package test_test

//revive:disable:dot-imports
import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/controlplane/server/mock"
	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler"
	otlpmock "github.com/neblic/platform/sampler/internal/sample/exporter/otlp/mock"
	internalSampler "github.com/neblic/platform/sampler/internal/sampler"
	"github.com/neblic/platform/sampler/sample"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestSampler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Sampler")
}

type nativeSample struct {
	ID int
}

func sendSamplerConfigHandler(samplerConfig *protos.SamplerConfig) func(protos.ControlPlane_SamplerConnServer) error {
	return func(stream protos.ControlPlane_SamplerConnServer) error {
		err := stream.Send(&protos.ServerToSampler{
			Message: &protos.ServerToSampler_ConfReq{
				ConfReq: &protos.ServerSamplerConfReq{
					SamplerConfig: samplerConfig,
				},
			},
		})
		Expect(err).ToNot(HaveOccurred())

		samplerConfRes, err := stream.Recv()
		Expect(err).ToNot(HaveOccurred())
		Expect(reflect.TypeOf(samplerConfRes.GetMessage())).
			To(Equal(reflect.TypeOf(&protos.SamplerToServer_ConfRes{})))

		return nil
	}
}

var _ = Describe("Sampler", func() {
	var (
		logger *logging.ZapLogger

		controlPlaneServer *mock.ControlPlaneServer

		logsReceiverLn net.Listener
		receiver       *otlpmock.LogsReceiver
	)

	BeforeEach(func() {
		var err error

		logger, err = logging.NewZapDev()
		Expect(err).ToNot(HaveOccurred())

		controlPlaneServer = mock.NewControlPlaneServer(GinkgoT())

		// initialize a mock log receiver (internally samples are converted to logs and sent to this receiver)
		logsReceiverLn, err = net.Listen("tcp", "localhost:")
		Expect(err).ToNot(HaveOccurred())
		receiver = otlpmock.OtlpLogsReceiverOnGRPCServer(logsReceiverLn)
	})

	AfterEach(func() {
		controlPlaneServer.Stop()
	})

	Describe("Sampler initialization", func() {
		When("default initial configuration is enabled", func() {
			It("should create default stream and digest configurations", func() {
				controlPlaneServer.SetSamplerHandlers(mock.RegisterSamplerHandler)
				controlPlaneServer.Start(GinkgoT())

				// initialize and start a sampler provider
				settings := sampler.Settings{
					ResourceName:      "sampled_service",
					ControlServerAddr: controlPlaneServer.Addr(),
					DataServerAddr:    logsReceiverLn.Addr().String(),
				}
				provider, err := sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
				Expect(err).ToNot(HaveOccurred())

				// create a sampler
				s, err := provider.Sampler("sampler1", sample.DynamicSchema{})
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has recieived the first configuration (with the default stream and digest)
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 1
					},
					time.Second, time.Millisecond*5,
				)

				sampled := s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, ""))
				Expect(sampled).To(BeTrue())

				Expect(s.Close()).ToNot(HaveOccurred())
			})
		})

		When("default initial configuration is disabled", func() {
			It("should not create any stream nor digest configuration", func() {
				controlPlaneServer.SetSamplerHandlers(mock.RegisterSamplerHandler)
				controlPlaneServer.Start(GinkgoT())

				// initialize and start a sampler provider
				settings := sampler.Settings{
					ResourceName:      "sampled_service",
					ControlServerAddr: controlPlaneServer.Addr(),
					DataServerAddr:    logsReceiverLn.Addr().String(),
				}
				provider, err := sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
				Expect(err).ToNot(HaveOccurred())

				// create a sampler
				s, err := provider.Sampler("sampler1", sample.DynamicSchema{}, sampler.WithoutDefaultInitialConfig())
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty)
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 1
					},
					time.Second, time.Millisecond*5,
				)

				sampled := s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, ""))
				Expect(sampled).To(BeFalse())

				Expect(s.Close()).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Not exporting samples", func() {
		When("there isn't a matching rule", func() {
			It("should not export samples", func() {
				controlPlaneServer.SetSamplerHandlers(
					mock.RegisterSamplerHandler,
					sendSamplerConfigHandler(
						&protos.SamplerConfig{
							Streams: []*protos.Stream{
								{
									Uid: uuid.NewString(),
									Rule: &protos.Rule{
										Language: protos.Rule_CEL, Expression: "sample.id==2",
									},
								},
							},
						}),
				)
				controlPlaneServer.Start(GinkgoT())

				// initialize and start a sampler provider
				settings := sampler.Settings{
					ResourceName:      "sampled_service",
					ControlServerAddr: controlPlaneServer.Addr(),
					DataServerAddr:    logsReceiverLn.Addr().String(),
				}
				provider, err := sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
				Expect(err).ToNot(HaveOccurred())

				// create a sampler
				s, err := provider.Sampler("sampler1", sample.DynamicSchema{}, sampler.WithoutDefaultInitialConfig())
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				sampled := s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, ""))
				Expect(sampled).To(BeFalse())

				Expect(s.Close()).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Not exporting raw samples", func() {
		var (
			err      error
			provider sampler.Provider
		)

		BeforeEach(func() {
			// configure and run a control plane server that registers the sampler and sends a configuration
			controlPlaneServer.SetSamplerHandlers(
				mock.RegisterSamplerHandler,
				sendSamplerConfigHandler(
					&protos.SamplerConfig{
						Streams: []*protos.Stream{
							{
								Uid: uuid.NewString(),
								Rule: &protos.Rule{
									Language: protos.Rule_CEL, Expression: "sample.id==1",
								},
							},
						},
					}),
			)
			controlPlaneServer.Start(GinkgoT())

			// initialize and start a sampler provider
			settings := sampler.Settings{
				ResourceName:      "sampled_service",
				ControlServerAddr: controlPlaneServer.Addr(),
				DataServerAddr:    logsReceiverLn.Addr().String(),
			}
			provider, err = sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
			Expect(err).ToNot(HaveOccurred())
		})

		When("there is a matching rule but exporting raw samples is disabled", func() {
			It("should not export the sample", func() {
				// create a sampler
				s, err := provider.Sampler("sampler1", sample.DynamicSchema{}, sampler.WithoutDefaultInitialConfig())
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				// send samples to sampler
				require.Never(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, ""))
						return receiver.TotalItems.Load() >= 1
					},
					time.Millisecond*500, time.Millisecond)

				Expect(s.Close()).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Exporting raw samples", func() {
		var (
			err      error
			provider sampler.Provider
		)

		BeforeEach(func() {
			// configure and run a control plane server that registers the sampler and sends a configuration
			controlPlaneServer.SetSamplerHandlers(
				mock.RegisterSamplerHandler,
				sendSamplerConfigHandler(
					&protos.SamplerConfig{
						Streams: []*protos.Stream{
							{
								Uid: uuid.NewString(),
								Rule: &protos.Rule{
									Language: protos.Rule_CEL, Expression: "sample.id==1",
								},
								ExportRawSamples: true,
							},
							{
								Uid: uuid.NewString(),
								Rule: &protos.Rule{
									Language: protos.Rule_CEL, Expression: "sample.ID==1",
								},
								ExportRawSamples: true,
							},
							{
								Uid: uuid.NewString(),
								Rule: &protos.Rule{
									Language: protos.Rule_CEL, Expression: `sample.sampler_uid == "1"`,
								},
								ExportRawSamples: true,
							},
						},
					}),
			)
			controlPlaneServer.Start(GinkgoT())

			// initialize and start a sampler provider
			settings := sampler.Settings{
				ResourceName:      "sampled_service",
				ControlServerAddr: controlPlaneServer.Addr(),
				DataServerAddr:    logsReceiverLn.Addr().String(),
			}
			provider, err = sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
			Expect(err).ToNot(HaveOccurred())
		})

		When("there is a matching rule and", func() {
			It("is a JSON sample it should export the sample", func() {
				// create a sampler
				s, err := provider.Sampler("sampler1", sample.DynamicSchema{}, sampler.WithoutDefaultInitialConfig())
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				// send samples to sampler
				sampled := s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, ""))
				Expect(sampled).To(BeTrue())

				// the receiver should have received the sample
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == 1
					},
					time.Second, time.Millisecond*5)

				Expect(s.Close()).ToNot(HaveOccurred())
			})

			It("is a native sample it should export the sample", func() {
				// create a sampler
				s, err := provider.Sampler("sampler1", sample.DynamicSchema{})
				Expect(err).ToNot(HaveOccurred())

				// wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				// send sample
				sampled := s.Sample(context.Background(), sample.NativeSample(nativeSample{ID: 1}, ""))
				Expect(sampled).To(BeTrue())

				// the receiver should have received the sample
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == 1
					},
					time.Second, time.Millisecond*5)

				Expect(s.Close()).ToNot(HaveOccurred())
			})

			It("is a proto sample it should export the sample", func() {
				// create a sampler
				s, err := provider.Sampler("sampler1", sample.NewProtoSchema(&protos.SamplerToServer{}))
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				// send sample
				sampled := s.Sample(context.Background(), sample.ProtoSample(&protos.SamplerToServer{SamplerUid: "1"}, ""))
				Expect(sampled).To(BeTrue())

				// wait until the receiver has received the sample
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == 1
					},
					time.Second, time.Millisecond)

				Expect(s.Close()).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Exporting digests", func() {
		var settings sampler.Settings

		BeforeEach(func() {
			// configure and run a control plane server that registers the sampler and sends a configuration
			streamUID := uuid.NewString()
			controlPlaneServer.SetSamplerHandlers(
				mock.RegisterSamplerHandler,
				sendSamplerConfigHandler(&protos.SamplerConfig{
					Streams: []*protos.Stream{
						{
							Uid: streamUID,
							Rule: &protos.Rule{
								Language: protos.Rule_CEL, Expression: "sample.id==1",
							},
						},
					},
					Digests: []*protos.Digest{
						{
							Uid:         uuid.NewString(),
							StreamUid:   streamUID,
							FlushPeriod: durationpb.New(200 * time.Millisecond),
							Type: &protos.Digest_St_{
								St: &protos.Digest_St{},
							},
						},
					},
				}),
			)
			controlPlaneServer.Start(GinkgoT())

			// common provider settings
			settings = sampler.Settings{
				ResourceName:      "sampled_service",
				ControlServerAddr: controlPlaneServer.Addr(),
				DataServerAddr:    logsReceiverLn.Addr().String(),
			}
		})

		When("there is a structure digest with sampler computation location", func() {
			It("should export structure digest samples", func() {
				providerLimitedOut, err := sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
				Expect(err).ToNot(HaveOccurred())

				// create a sampler
				s, err := providerLimitedOut.Sampler("sampler1", sample.DynamicSchema{},
					sampler.WithInitialLimiterOutLimit(10),
					sampler.WithoutDefaultInitialConfig(),
					sampler.WithInitialStructDigest(control.ComputationLocationSampler),
				)
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				// send sample
				sampled := s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, ""))
				Expect(sampled).To(BeTrue())

				// wait until the receiver has received the digest
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						totalItems := receiver.TotalItems.Load()
						return totalItems >= 1
					},
					time.Second, time.Millisecond)

				// check it is a digest
				// TODO: instead of hardcoding this check, we could create a public function
				// that converts otlp logs back to Sample structs
				lastReq := receiver.GetLastRequest()
				require.Equal(GinkgoT(), lastReq.LogRecordCount(), 1)
				scopeLogs := lastReq.ResourceLogs().At(0).ScopeLogs()
				require.Equal(GinkgoT(), scopeLogs.Len(), 1)
				logRecords := scopeLogs.At(0).LogRecords()
				require.Equal(GinkgoT(), logRecords.Len(), 1)
				sampleTypeVal, ok := logRecords.At(0).Attributes().Get("sample_type")
				require.True(GinkgoT(), ok)
				assert.Equal(GinkgoT(), sampleTypeVal.AsString(), "struct-digest")
			})
		})

		When("there is a structure digest with collector computation location", func() {
			It("should never export structure digest samples", func() {
				providerLimitedOut, err := sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
				Expect(err).ToNot(HaveOccurred())

				// create a sampler
				s, err := providerLimitedOut.Sampler("sampler1", sample.DynamicSchema{},
					sampler.WithInitialLimiterOutLimit(10),
					sampler.WithoutDefaultInitialConfig(),
					sampler.WithInitialStructDigest(control.ComputationLocationCollector),
				)
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				// send sample
				sampled := s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, ""))
				Expect(sampled).To(BeTrue())

				// no digest has to be generated
				require.Never(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return receiver.TotalItems.Load() == 1
					},
					time.Millisecond*500, time.Millisecond)
			})
		})
	})

	Describe("Limiting exported samples", func() {
		var settings sampler.Settings

		BeforeEach(func() {
			// configure and run a control plane server that registers the sampler and sends a configuration
			controlPlaneServer.SetSamplerHandlers(
				mock.RegisterSamplerHandler,
				sendSamplerConfigHandler(&protos.SamplerConfig{
					Streams: []*protos.Stream{
						{
							Uid: uuid.NewString(),
							Rule: &protos.Rule{
								Language: protos.Rule_CEL, Expression: "sample.id==1",
							},
							ExportRawSamples: true,
						},
					},
				}),
			)
			controlPlaneServer.Start(GinkgoT())

			// common provider settings
			settings = sampler.Settings{
				ResourceName:      "sampled_service",
				ControlServerAddr: controlPlaneServer.Addr(),
				DataServerAddr:    logsReceiverLn.Addr().String(),
			}
		})

		When("there is an out limiter set", func() {
			It("should not export more samples than the allowed by the limiter settings", func() {
				providerLimitedOut, err := sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
				Expect(err).ToNot(HaveOccurred())

				// create a sampler
				s, err := providerLimitedOut.Sampler("sampler1", sample.DynamicSchema{},
					sampler.WithInitialLimiterOutLimit(10),
					sampler.WithoutDefaultInitialConfig(),
				)
				Expect(err).ToNot(HaveOccurred())

				// wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				// send a large amount of samples so the limiter kicks in
				numSampled := 0
				for i := 0; i < 1000; i++ {
					sampled := s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, ""))

					if sampled {
						numSampled++
					}
				}

				Expect(numSampled).To(Equal(10))

				// the receiver should have received the sample
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == int32(numSampled)
					},
					time.Second, time.Millisecond)

				Expect(s.Close()).ToNot(HaveOccurred())
			})
		})

		When("there is an in limiter set", func() {
			It("should not export more samples than the allowed by the limiter settings", func() {
				providerLimitedIn, err := sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
				Expect(err).ToNot(HaveOccurred())

				// create a sampler
				s, err := providerLimitedIn.Sampler("sampler1", sample.DynamicSchema{},
					sampler.WithInitialLimiterInLimit(5),
					sampler.WithoutDefaultInitialConfig(),
				)
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				// send a large amount of samples so the limiter kicks in
				numSampled := 0
				for i := 0; i < 1000; i++ {
					sampled := s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, ""))

					if sampled {
						numSampled++
					}
				}

				Expect(numSampled).To(Equal(5))

				// the receiver should have received the sample
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == int32(numSampled)
					},
					time.Second, time.Millisecond)

				Expect(s.Close()).ToNot(HaveOccurred())
			})

		})

		When("there is an in sampler set", func() {
			It("should not export samples if their determinant is not selected", func() {
				providerSampledIn, err := sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
				Expect(err).ToNot(HaveOccurred())
				// create a sampler
				s, err := providerSampledIn.Sampler("sampler1", sample.DynamicSchema{},
					sampler.WithInitialLimiterInLimit(1000),
					sampler.WithInitialDeterministicSamplingIn(2),
					sampler.WithInitialLimiterOutLimit(1000),
					sampler.WithoutDefaultInitialConfig(),
				)
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				// should not be sampled
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						sampled := s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, "some_non_matching_key"))
						return !sampled
					},
					time.Second, time.Millisecond)

				// should all be sampled
				numSampled := 0
				for i := 0; i < 100; i++ {
					sampled := s.Sample(context.Background(), sample.JSONSample(`{"id": 1}`, "some_matching_key"))

					if sampled {
						numSampled++
					}
				}

				Expect(numSampled).To(Equal(100))

				// the receiver should have received all the samples
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == int32(numSampled)
					},
					time.Second, time.Millisecond)

				Expect(s.Close()).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Sending stats", func() {
		When("the sampler is running", func() {
			It("should send stats periodically", func() {
				statsReceived := make(chan struct{})
				controlPlaneServer.SetSamplerHandlers(
					mock.RegisterSamplerHandler,
					func(stream protos.ControlPlane_SamplerConnServer) error {
						stats, err := stream.Recv()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(reflect.TypeOf(stats.GetMessage())).
							To(Equal(reflect.TypeOf(&protos.SamplerToServer_SamplerStatsMsg{})))

						statsReceived <- struct{}{}
						return nil
					},
				)
				controlPlaneServer.Start(GinkgoT())

				// initialize and start a sampler provider
				settings := sampler.Settings{
					ResourceName:      "sampled_service",
					ControlServerAddr: controlPlaneServer.Addr(),
					DataServerAddr:    logsReceiverLn.Addr().String(),
				}
				provider, err := sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
				Expect(err).ToNot(HaveOccurred())

				// create a sampler
				s, err := provider.Sampler("sampler1", sample.DynamicSchema{},
					sampler.WithUpdateStatsPeriod(time.Second),
					sampler.WithoutDefaultInitialConfig(),
				)
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty)
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 1
					},
					time.Second, time.Millisecond*5,
				)

				<-statsReceived

				Expect(s.Close()).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Forwarding errors", func() {
		When("an error channel is provided", func() {
			It("should forward errors to the channel", func() {

				// start control plane server
				controlPlaneServer.SetSamplerHandlers(
					mock.RegisterSamplerHandler,
					sendSamplerConfigHandler(&protos.SamplerConfig{
						Streams: []*protos.Stream{
							{
								Uid: uuid.NewString(),
								Rule: &protos.Rule{
									Language: protos.Rule_CEL, Expression: "sample.id==1",
								},
							},
						},
					}),
				)
				controlPlaneServer.Start(GinkgoT())

				// initialize and start a sampler provider
				settings := sampler.Settings{
					ResourceName:      "sampled_service",
					ControlServerAddr: controlPlaneServer.Addr(),
					DataServerAddr:    logsReceiverLn.Addr().String(),
				}
				errCh := make(chan error, 1)

				provider, err := sampler.NewProvider(context.Background(), settings,
					sampler.WithLogger(logger),
					sampler.WithSamplerErrorChannel(errCh),
				)
				Expect(err).ToNot(HaveOccurred())

				// create a sampler
				s, err := provider.Sampler("sampler1", sample.DynamicSchema{},
					sampler.WithUpdateStatsPeriod(time.Second),
					sampler.WithoutDefaultInitialConfig(),
				)
				Expect(err).ToNot(HaveOccurred())

				// Wait until sampler has received the initial configuration (empty) and the posterior update
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						return s.(*internalSampler.Sampler).ConfigUpdates() == 2
					},
					time.Second, time.Millisecond*5,
				)

				// send an invalid sample
				sampled := s.Sample(context.Background(), sample.JSONSample(`invalid_json: `, ""))
				Expect(sampled).To(BeFalse())

				/// expect an error to be received
				require.Eventually(GinkgoT(),
					func() bool {
						<-errCh
						return true
					},
					time.Second, time.Millisecond)

				Expect(s.Close()).ToNot(HaveOccurred())
			})
		})
	})
})

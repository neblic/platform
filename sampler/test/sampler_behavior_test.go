package test_test

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/controlplane/server/mock"
	"github.com/neblic/platform/logging"
	"github.com/neblic/platform/sampler"
	"github.com/neblic/platform/sampler/defs"
	otlpmock "github.com/neblic/platform/sampler/internal/sample/exporter/otlp/mock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
)

func TestSampler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Sampler")
}

type nativeSample struct {
	ID int
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

	Describe("Not exporting samples", func() {
		When("there isn't a matching rule", func() {
			It("should not export samples", func() {
				registered := make(chan struct{})
				controlPlaneServer.SetSamplerHandlers(
					mock.RegisterSamplerHandler,
					func(stream protos.ControlPlane_SamplerConnServer) error {
						registered <- struct{}{}
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
				p, err := provider.Sampler("sampler1", defs.DynamicSchema{})
				Expect(err).ToNot(HaveOccurred())

				sampled := p.Sample(context.Background(), defs.JsonSample(`{"id": 1}`, ""))
				Expect(sampled).To(BeFalse())

				Expect(p.Close()).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Exporting samples", func() {
		var (
			err      error
			provider defs.Provider
		)
		configured := make(chan struct{})

		BeforeEach(func() {
			// configure and run a control plane server that registers the sampler and sends a configuration
			controlPlaneServer.SetSamplerHandlers(
				mock.RegisterSamplerHandler,
				func(stream protos.ControlPlane_SamplerConnServer) error {
					err = stream.Send(&protos.ServerToSampler{
						Message: &protos.ServerToSampler_ConfReq{
							ConfReq: &protos.ServerSamplerConfReq{
								SamplerConfig: &protos.SamplerConfig{
									Streams: []*protos.Stream{
										{
											Uid: uuid.NewString(),
											Rule: &protos.Stream_Rule{
												Language: protos.Stream_Rule_CEL, Rule: "sample.id==1",
											},
										},
										{
											Uid: uuid.NewString(),
											Rule: &protos.Stream_Rule{
												Language: protos.Stream_Rule_CEL, Rule: "sample.ID==1",
											},
										},
										{
											Uid: uuid.NewString(),
											Rule: &protos.Stream_Rule{
												Language: protos.Stream_Rule_CEL, Rule: `sample.sampler_uid == "1"`,
											},
										},
									},
								},
							},
						},
					})
					Expect(err).ToNot(HaveOccurred())

					samplerConfRes, err := stream.Recv()
					Expect(err).ToNot(HaveOccurred())
					Expect(reflect.TypeOf(samplerConfRes.GetMessage())).
						To(Equal(reflect.TypeOf(&protos.SamplerToServer_ConfRes{})))

					configured <- struct{}{}
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
			provider, err = sampler.NewProvider(context.Background(), settings, sampler.WithLogger(logger))
			Expect(err).ToNot(HaveOccurred())

		})

		When("there is a matching rule and", func() {
			It("is a JSON sample it should export the sample", func() {
				// create a sampler
				p, err := provider.Sampler("sampler1", defs.DynamicSchema{})
				Expect(err).ToNot(HaveOccurred())

				// wait until the server has configured the sampler
				<-configured

				// send samples to sampler until it is sampled
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						sampled := p.Sample(context.Background(), defs.JsonSample(`{"id": 1}`, ""))
						return sampled
					},
					time.Second, time.Millisecond)

				// the receiver should have received the sample
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == 1
					},
					time.Second, time.Millisecond)

				Expect(p.Close()).ToNot(HaveOccurred())
			})

			It("is a native sample it should export the sample", func() {
				// create a sampler
				p, err := provider.Sampler("sampler1", defs.DynamicSchema{})
				Expect(err).ToNot(HaveOccurred())

				// wait until the server has configured the sampler
				<-configured

				// send samples to sampler until it is sampled
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						sampled := p.Sample(context.Background(), defs.NativeSample(nativeSample{ID: 1}, ""))
						Expect(err).ToNot(HaveOccurred())
						return sampled
					},
					time.Second, time.Millisecond)

				// the receiver should have received the sample
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == 1
					},
					time.Second, time.Millisecond)

				Expect(p.Close()).ToNot(HaveOccurred())
			})

			It("is a proto sample it should export the sample", func() {
				// create a sampler
				p, err := provider.Sampler("sampler1", defs.NewProtoSchema(&protos.SamplerToServer{}))
				Expect(err).ToNot(HaveOccurred())

				// wait until the server has configured the sampler
				<-configured

				// send samples to sampler until it is sampled
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						sampled := p.Sample(context.Background(), defs.ProtoSample(&protos.SamplerToServer{SamplerUid: "1"}, ""))
						Expect(err).ToNot(HaveOccurred())
						return sampled
					},
					time.Second, time.Millisecond)

				// the receiver should have received the sample
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == 1
					},
					time.Second, time.Millisecond)

				Expect(p.Close()).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Limiting exported samples", func() {
		var (
			err                                                      error
			providerLimitedOut, providerLimitedIn, providerSampledIn defs.Provider
		)
		configured := make(chan struct{})

		BeforeEach(func() {
			// configure and run a control plane server that registers the sampler and sends a configuration
			controlPlaneServer.SetSamplerHandlers(
				mock.RegisterSamplerHandler,
				func(stream protos.ControlPlane_SamplerConnServer) error {
					err = stream.Send(&protos.ServerToSampler{
						Message: &protos.ServerToSampler_ConfReq{
							ConfReq: &protos.ServerSamplerConfReq{
								SamplerConfig: &protos.SamplerConfig{
									Streams: []*protos.Stream{
										{
											Uid: uuid.NewString(),
											Rule: &protos.Stream_Rule{
												Language: protos.Stream_Rule_CEL, Rule: "sample.id==1",
											},
										},
									},
								},
							},
						},
					})
					Expect(err).ToNot(HaveOccurred())

					samplerConfRes, err := stream.Recv()
					Expect(err).ToNot(HaveOccurred())
					Expect(reflect.TypeOf(samplerConfRes.GetMessage())).
						To(Equal(reflect.TypeOf(&protos.SamplerToServer_ConfRes{})))

					configured <- struct{}{}
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

			providerLimitedIn, err = sampler.NewProvider(context.Background(), settings,
				sampler.WithLimiterInLimit(5),
				sampler.WithLogger(logger),
			)
			Expect(err).ToNot(HaveOccurred())

			providerSampledIn, err = sampler.NewProvider(context.Background(), settings,
				sampler.WithLimiterInLimit(1000),
				sampler.WithDeterministicSamplingIn(2),
				sampler.WithLimiterOutLimit(1000),
				sampler.WithLogger(logger),
			)
			Expect(err).ToNot(HaveOccurred())

			providerLimitedOut, err = sampler.NewProvider(context.Background(), settings,
				sampler.WithLimiterOutLimit(10),
				sampler.WithLogger(logger),
			)
			Expect(err).ToNot(HaveOccurred())
		})

		When("there is an out limiter set", func() {
			It("should not export more samples than the allowed by the limiter settings", func() {
				// create a sampler
				p, err := providerLimitedOut.Sampler("sampler1", defs.DynamicSchema{})
				Expect(err).ToNot(HaveOccurred())

				// wait until the server has configured the sampler
				<-configured

				// send samples to sampler until it is sampled
				// we do this so we are sure the config has been read and applied by the sampler
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						sampled := p.Sample(context.Background(), defs.JsonSample(`{"id": 1}`, ""))
						return sampled
					},
					time.Second, time.Millisecond)

				// send a large amount of samples so the limiter kicks in
				numSampled := 0
				for i := 0; i < 1000; i++ {
					sampled := p.Sample(context.Background(), defs.JsonSample(`{"id": 1}`, ""))

					if sampled {
						numSampled++
					}
				}

				Expect(numSampled + 1).To(Equal(10))

				// the receiver should have received the sample
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == int32(numSampled+1)
					},
					time.Second, time.Millisecond)

				Expect(p.Close()).ToNot(HaveOccurred())
			})
		})

		When("there is an in limiter set", func() {
			It("should not export more samples than the allowed by the limiter settings", func() {
				// create a sampler
				p, err := providerLimitedIn.Sampler("sampler1", defs.DynamicSchema{})
				Expect(err).ToNot(HaveOccurred())

				// wait until the server has configured the sampler
				<-configured

				// send samples to sampler until it is sampled
				// we do this so we are sure the config has been read and applied by the sampler
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						sampled := p.Sample(context.Background(), defs.JsonSample(`{"id": 1}`, ""))
						return sampled
					},
					time.Second, time.Millisecond)

				// send a large amount of samples so the limiter kicks in
				numSampled := 0
				for i := 0; i < 1000; i++ {
					sampled := p.Sample(context.Background(), defs.JsonSample(`{"id": 1}`, ""))

					if sampled {
						numSampled++
					}
				}

				Expect(numSampled + 1).To(Equal(5))

				// the receiver should have received the sample
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == int32(numSampled+1)
					},
					time.Second, time.Millisecond)

				Expect(p.Close()).ToNot(HaveOccurred())
			})

		})

		When("there is an in sampler set", func() {
			It("should not export samples if their determinant is not selected", func() {
				// create a sampler
				p, err := providerSampledIn.Sampler("sampler1", defs.DynamicSchema{})
				Expect(err).ToNot(HaveOccurred())

				// wait until the server has configured the sampler
				<-configured

				// send samples to sampler until it is sampled
				// we do this so we are sure the config has been read and applied by the sampler
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						sampled := p.Sample(context.Background(), defs.JsonSample(`{"id": 1}`, "some_matching_key"))
						return sampled
					},
					time.Second, time.Millisecond)

				// should not be sampled
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()

						sampled := p.Sample(context.Background(), defs.JsonSample(`{"id": 1}`, "some_non_matching_key"))
						return !sampled
					},
					time.Second, time.Millisecond)

				// should all be sampled
				numSampled := 0
				for i := 0; i < 100; i++ {
					sampled := p.Sample(context.Background(), defs.JsonSample(`{"id": 1}`, "some_matching_key"))

					if sampled {
						numSampled++
					}
				}

				Expect(numSampled).To(Equal(100))

				// the receiver should have received all the samples
				require.Eventually(GinkgoT(),
					func() bool {
						defer GinkgoRecover()
						return receiver.TotalItems.Load() == int32(numSampled+1)
					},
					time.Second, time.Millisecond)

				Expect(p.Close()).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Sending stats", func() {
		When("the sampler is running", func() {
			It("should send stats periodically", func() {
				registered := make(chan struct{})
				statsReceived := make(chan struct{})
				controlPlaneServer.SetSamplerHandlers(
					mock.RegisterSamplerHandler,
					func(stream protos.ControlPlane_SamplerConnServer) error {
						registered <- struct{}{}
						return nil
					},
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
				provider, err := sampler.NewProvider(context.Background(), settings,
					sampler.WithUpdateStatsPeriod(time.Second),
					sampler.WithLogger(logger),
				)
				Expect(err).ToNot(HaveOccurred())

				// create a sampler
				p, err := provider.Sampler("sampler1", defs.DynamicSchema{})
				Expect(err).ToNot(HaveOccurred())

				<-registered
				<-statsReceived

				Expect(p.Close()).ToNot(HaveOccurred())
			})
		})
	})
})

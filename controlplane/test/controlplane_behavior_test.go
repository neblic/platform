package test

//revive:disable:dot-imports
import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/client"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/internal/test"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/controlplane/sampler"
	"github.com/neblic/platform/controlplane/server"
	"github.com/neblic/platform/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const condTimeout = time.Duration(1) * time.Second

func TestControlPlane(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Control Plane")
}

func waitClientRegistered(c *client.Client) chan struct{} {
	registered := make(chan struct{})
	states := c.StateChanges()

	go func() {
		defer GinkgoRecover()

		// check first if it is already registered to avoid missing the event
		if c.State() == client.Registered {
			registered <- struct{}{}
			return
		}

	loop:
		for {
			state, more := <-states
			if !more {
				break loop
			}

			if state == client.Registered {
				registered <- struct{}{}

				// do not break since next events need to be consumed to avoid blocking the client
			}
		}
	}()

	return registered
}

func waitSamplerRegistered(p *sampler.Sampler) chan struct{} {
	registered := make(chan struct{})
	events := p.Events()

	go func() {
		defer GinkgoRecover()

		// check first if it is already registered to avoid missing the event
		if p.State() == sampler.Registered {
			registered <- struct{}{}
			return
		}

	loop:
		for {
			event, more := <-events
			if !more {
				break loop
			}

			stateUpdate, ok := event.(sampler.StateUpdate)
			if ok && stateUpdate.State == sampler.Registered {
				registered <- struct{}{}

				// do not break since next events need to be consumed to avoid blocking the sampler
			}
		}
	}()

	return registered
}

var _ = Describe("ControlPlane", func() {
	Describe("Encrypted connection", func() {
		var (
			logger logging.Logger
			s      *server.Server
		)

		BeforeEach(func() {
			var err error

			logger, err = logging.NewZapDev()
			Expect(err).ToNot(HaveOccurred())

			opts := []server.Option{
				server.WithLogger(logger),
				server.WithTLS("./assets/localhost.crt", "./assets/localhost.key"),
			}

			s, err = server.New("server_uid", opts...)
			Expect(err).ToNot(HaveOccurred())

			err = s.Start("localhost:")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := s.Stop(condTimeout)
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("When client connects to a TLS enabled server", func() {
			It("should register with server automatically", func() {
				c := client.New(uuid.New().String(),
					client.WithTLS(),
					client.WithTLSCACert("./assets/localhost.crt"),
					client.WithLogger(logger))

				err := c.Connect(s.Addr().String())
				registered := waitClientRegistered(c)

				<-registered
				Expect(err).ToNot(HaveOccurred())

				Expect(c.Close(condTimeout)).ToNot(HaveOccurred())
			})
		})

		Describe("When sampler connects to a TLS enabled server", func() {
			It("should register with server automatically", func() {
				p := sampler.New("sampler1", "resource1",
					sampler.WithTLS(),
					sampler.WithTLSCACert("./assets/localhost.crt"),
					sampler.WithLogger(logger))

				err := p.Connect(s.Addr().String())
				registered := waitSamplerRegistered(p)

				<-registered
				Expect(err).ToNot(HaveOccurred())

				Expect(p.Close(condTimeout)).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Encrypted connection with auth", func() {
		var (
			logger logging.Logger
			s      *server.Server
		)

		BeforeEach(func() {
			var err error

			logger, err = logging.NewZapDev()
			Expect(err).ToNot(HaveOccurred())

			opts := []server.Option{
				server.WithLogger(logger),
				server.WithTLS("./assets/localhost.crt", "./assets/localhost.key"),
				server.WithAuthBearer("some_token"),
			}

			s, err = server.New("server_uid", opts...)
			Expect(err).ToNot(HaveOccurred())

			err = s.Start("localhost:")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := s.Stop(condTimeout)
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("When client connects to a TLS enabled server with bearer auth", func() {
			It("should register with server automatically", func() {
				c := client.New(uuid.New().String(),
					client.WithTLS(),
					client.WithTLSCACert("./assets/localhost.crt"),
					client.WithLogger(logger),
					client.WithAuthBearer("some_token"),
				)

				err := c.Connect(s.Addr().String())
				registered := waitClientRegistered(c)

				<-registered
				Expect(err).ToNot(HaveOccurred())

				Expect(c.Close(condTimeout)).ToNot(HaveOccurred())
			})
		})

		Describe("When sampler connects to a TLS enabled server with bearer auth", func() {
			It("should register with server automatically", func() {
				p := sampler.New("sampler1", "resource1",
					sampler.WithTLS(),
					sampler.WithTLSCACert("./assets/localhost.crt"),
					sampler.WithLogger(logger),
					sampler.WithAuthBearer("some_token"),
				)

				err := p.Connect(s.Addr().String())
				registered := waitSamplerRegistered(p)

				<-registered
				Expect(err).ToNot(HaveOccurred())

				Expect(p.Close(condTimeout)).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Unencrypted connection", func() {
		var (
			logger logging.Logger
			s      *server.Server
		)

		BeforeEach(func() {
			var err error

			logger, err = logging.NewZapDev()
			Expect(err).ToNot(HaveOccurred())

			s, err = server.New("server_uid", server.WithLogger(logger))
			Expect(err).ToNot(HaveOccurred())

			err = s.Start("localhost:")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := s.Stop(condTimeout)
			Expect(err).ToNot(HaveOccurred())
		})

		// Client -> Server
		Describe("Client -> Server", func() {
			// 1. Client registration
			Describe("When client connects", func() {
				It("should register with server automatically", func() {
					c := client.New(uuid.New().String(), client.WithLogger(logger))
					registered := waitClientRegistered(c)

					err := c.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					<-registered
					Expect(c.Close(condTimeout)).ToNot(HaveOccurred())
				})
			})
		})

		// sampler -> Server
		Describe("sampler -> Server", func() {
			// 1. sampler registration
			Describe("When sampler connects", func() {
				It("should register with server automatically", func() {
					p := sampler.New("sampler1", "resource1", sampler.WithLogger(logger))
					registered := waitSamplerRegistered(p)

					err := p.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					<-registered
					Expect(p.Close(condTimeout)).ToNot(HaveOccurred())
				})

				It("should send all its information", func() {
					// fill all initialConfig fields with demo data
					initialConfig := control.NewSamplerConfigUpdate()
					initialStream := control.Stream{
						UID:  control.SamplerStreamUID(uuid.NewString()),
						Name: "some_stream_name",
						StreamRule: control.Rule{
							Lang:       control.NewRuleLangFromProto(protos.Rule_CEL),
							Expression: "some_CEL_rule",
						},
					}
					initialConfig.StreamUpdates = []control.StreamUpdate{
						{
							Op:     control.StreamUpsert,
							Stream: initialStream,
						},
					}

					initialConfig.LimiterIn = &control.LimiterConfig{Limit: 100}
					initialConfig.LimiterOut = &control.LimiterConfig{Limit: 10}
					initialConfig.SamplingIn = &control.SamplingConfig{
						SamplingType: control.DeterministicSamplingType,
						DeterministicSampling: control.DeterministicSamplingConfig{
							SampleRate:             100,
							SampleEmptyDeterminant: true,
						},
					}

					initialDigest := control.Digest{
						UID:                 control.SamplerDigestUID(uuid.NewString()),
						Name:                "some_digest_name",
						StreamUID:           initialStream.UID,
						FlushPeriod:         time.Second * time.Duration(60),
						ComputationLocation: control.ComputationLocationSampler,
						Type:                control.DigestTypeValue,
						Value: &control.DigestValue{
							MaxProcessedFields: int(100),
						},
					}
					initialConfig.DigestUpdates = []control.DigestUpdate{
						{
							Op:     control.DigestUpsert,
							Digest: initialDigest,
						},
					}

					initialEvent := control.Event{
						UID:        control.SamplerEventUID(uuid.NewString()),
						Name:       "some_event_name",
						StreamUID:  initialStream.UID,
						SampleType: control.RawSampleType,
						Rule: control.Rule{
							Lang:       control.NewRuleLangFromProto(protos.Rule_CEL),
							Expression: "some_CEL_rule",
						},
						Limiter: control.LimiterConfig{
							Limit: 100,
						},
					}
					initialConfig.EventUpdates = []control.EventUpdate{
						{
							Op:    control.EventUpsert,
							Event: initialEvent,
						},
					}

					tags := control.Tags{{Name: "some_tag_name", Attrs: map[string]string{"some_attr_name": "some_attr_value"}}}
					p := sampler.New("sampler1", "resource1",
						sampler.WithLogger(logger),
						sampler.WithTags(tags...),
						sampler.WithInitialConfig(initialConfig),
					)
					registered := waitSamplerRegistered(p)

					err := p.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					<-registered

					samplers, err := s.GetSamplers("resource1", "sampler1")
					Expect(err).ToNot(HaveOccurred())
					Expect(samplers).To(HaveLen(1))

					// we use cmp.Diff since it provides better error messages
					diff := cmp.Diff(samplers[0].Config, control.SamplerConfig{
						Streams: map[control.SamplerStreamUID]control.Stream{
							initialStream.UID: initialStream,
						},
						LimiterIn:  &control.LimiterConfig{Limit: 100},
						LimiterOut: &control.LimiterConfig{Limit: 10},
						SamplingIn: &control.SamplingConfig{
							SamplingType: control.DeterministicSamplingType,
							DeterministicSampling: control.DeterministicSamplingConfig{
								SampleRate:             100,
								SampleEmptyDeterminant: true,
							},
						},
						Digests: map[control.SamplerDigestUID]control.Digest{
							initialDigest.UID: initialDigest,
						},
						Events: map[control.SamplerEventUID]control.Event{
							initialEvent.UID: initialEvent,
						},
					})
					Expect(diff).To(BeEmpty())
					Expect(samplers[0].Tags).To(Equal(tags))

					Expect(p.Close(condTimeout)).ToNot(HaveOccurred())
				})
			})
		})

		// Client -> Server -> sampler
		Describe("Client -> Server -> sampler", func() {
			// 1. sampler configuration
			Describe("When client sends a configuration by sampler name", func() {
				It("should be forwarded to the sampler", func() {
					c := client.New(uuid.New().String(), client.WithLogger(logger))
					clientRegistered := waitClientRegistered(c)
					err := c.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					p := sampler.New("sampler1", "resource1", sampler.WithLogger(logger))
					samplerRegistered := waitSamplerRegistered(p)
					err = p.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					<-clientRegistered
					<-samplerRegistered

					testStream := control.Stream{
						UID: "some_stream_uid",
						StreamRule: control.Rule{
							Lang:       control.NewRuleLangFromProto(protos.Rule_CEL),
							Expression: "some_CEL_rule",
						},
					}

					samplerConfigUpdate := &control.SamplerConfigUpdate{
						StreamUpdates: []control.StreamUpdate{
							{
								Op:     control.StreamUpsert,
								Stream: testStream,
							},
						},
					}
					err = c.ConfigureSampler(context.Background(), "resource1", p.Name(), samplerConfigUpdate)
					Expect(err).ToNot(HaveOccurred())

					test.AssertWithTimeout(
						func() bool { return len(p.Config().Streams) == 1 },
						condTimeout,
						func() {
							Expect(len(p.Config().Streams)).To(Equal(1))
							Expect(p.Config()).To(Equal(control.SamplerConfig{
								Streams: map[control.SamplerStreamUID]control.Stream{
									testStream.UID: testStream,
								},
							}))
						},
					)

					Expect(c.Close(condTimeout)).ToNot(HaveOccurred())
				})
			})
			Describe("When client sends a configuration by sampler id", func() {
				It("should be forwarded to the sampler", func() {
					c := client.New(uuid.New().String(), client.WithLogger(logger))
					clientRegistered := waitClientRegistered(c)
					err := c.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					p := sampler.New("sampler1", "resource1", sampler.WithLogger(logger))
					samplerRegistered := waitSamplerRegistered(p)
					err = p.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					<-clientRegistered
					<-samplerRegistered

					testStream := control.Stream{
						UID: "some_stream_uid",
						StreamRule: control.Rule{
							Lang:       control.NewRuleLangFromProto(protos.Rule_CEL),
							Expression: "some_CEL_rule",
						},
					}

					samplerConfigUpdate := &control.SamplerConfigUpdate{
						StreamUpdates: []control.StreamUpdate{
							{
								Op:     control.StreamUpsert,
								Stream: testStream,
							},
						},
					}

					err = c.ConfigureSampler(context.Background(), "resource1", p.Name(), samplerConfigUpdate)
					Expect(err).ToNot(HaveOccurred())

					test.AssertWithTimeout(
						func() bool { return len(p.Config().Streams) == 1 },
						condTimeout,
						func() {
							Expect(len(p.Config().Streams)).To(Equal(1))
							Expect(p.Config()).To(Equal(control.SamplerConfig{
								Streams: map[control.SamplerStreamUID]control.Stream{
									testStream.UID: testStream,
								},
							}))
						},
					)

					Expect(c.Close(condTimeout)).ToNot(HaveOccurred())
				})
			})

			// 2. List Samplers
			Describe("When client lists samplers", func() {
				It("should receive all registered samplers", func() {
					c := client.New(uuid.New().String(), client.WithLogger(logger))
					clientRegistered := waitClientRegistered(c)
					err := c.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					p1 := sampler.New("sampler1", "resource1", sampler.WithLogger(logger))
					sampler1Registered := waitSamplerRegistered(p1)
					err = p1.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					p2 := sampler.New("sampler2", "resource2", sampler.WithLogger(logger))
					sampler2Registered := waitSamplerRegistered(p2)
					err = p2.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					<-clientRegistered
					<-sampler1Registered
					<-sampler2Registered

					samplers, err := c.ListSamplers(context.Background())
					Expect(err).ToNot(HaveOccurred())

					Expect(len(samplers)).To(Equal(2))
					Expect(samplers[0].UID).To(BeElementOf(p1.UID(), p2.UID()))
					Expect(samplers[1].UID).To(BeElementOf(p1.UID(), p2.UID()))

					Expect(c.Close(condTimeout)).ToNot(HaveOccurred())
				})
			})

			// 3. Configuration recovery
			Describe("When sampler reconnects", func() {
				It("should recover the previous configuration", func() {
					c := client.New(uuid.New().String(), client.WithLogger(logger))
					clientRegistered := waitClientRegistered(c)
					err := c.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					p := sampler.New("sampler1", "resource1", sampler.WithLogger(logger))
					samplerRegistered := waitSamplerRegistered(p)
					err = p.Connect(s.Addr().String())
					Expect(err).ToNot(HaveOccurred())

					<-clientRegistered
					<-samplerRegistered

					testStream := control.Stream{
						UID: "some_stream_uid",
						StreamRule: control.Rule{
							Lang:       control.NewRuleLangFromProto(protos.Rule_CEL),
							Expression: "some_CEL_rule",
						},
					}

					samplerConfigUpdate := &control.SamplerConfigUpdate{
						StreamUpdates: []control.StreamUpdate{
							{
								Op:     control.StreamUpsert,
								Stream: testStream,
							},
						},
					}

					err = c.ConfigureSampler(context.Background(), "resource1", p.Name(), samplerConfigUpdate)
					Expect(err).ToNot(HaveOccurred())

					test.AssertWithTimeout(
						func() bool { return len(p.Config().Streams) == 1 },
						condTimeout,
						func() {
							Expect(len(p.Config().Streams)).To(Equal(1))
							Expect(p.Config()).To(Equal(control.SamplerConfig{
								Streams: map[control.SamplerStreamUID]control.Stream{
									testStream.UID: testStream,
								},
							}))
						},
					)

					Expect(p.Close(condTimeout)).ToNot(HaveOccurred())

					p2 := sampler.New("sampler1", "resource1", sampler.WithLogger(logger))
					samplerRegistered2 := waitSamplerRegistered(p2)
					err = p2.Connect(s.Addr().String())

					Expect(err).ToNot(HaveOccurred())

					<-samplerRegistered2

					test.AssertWithTimeout(
						func() bool { return len(p.Config().Streams) == 1 },
						condTimeout,
						func() {
							Expect(len(p.Config().Streams)).To(Equal(1))
							Expect(p.Config()).To(Equal(control.SamplerConfig{
								Streams: map[control.SamplerStreamUID]control.Stream{
									testStream.UID: testStream,
								},
							}))
						},
					)

					Expect(c.Close(condTimeout)).ToNot(HaveOccurred())
					Expect(p2.Close(condTimeout)).ToNot(HaveOccurred())
				})
			})

		})

		// TODO
		// sampler -> Server -> Client
		// 1. Stats forwarding
	})

})

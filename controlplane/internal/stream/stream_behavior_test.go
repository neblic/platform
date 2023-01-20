package stream_test

//revive:disable:dot-imports
import (
	"context"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/internal/stream"
	"github.com/neblic/platform/controlplane/internal/test"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/controlplane/server/mock"
	"github.com/neblic/platform/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
)

func TestStream(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stream Suite")
}

const condTimeout = time.Duration(100) * time.Millisecond

func registerClient(clientStream *stream.Stream[*protos.ServerToClient, *protos.ClientToServer], clientServerConn protos.ControlPlane_ClientConnServer) {
	// assert client is not registered
	Expect(clientStream.State().String()).To(Not(Equal(stream.Registered.String())))

	// assert it has started registering
	msg, err := clientServerConn.Recv()
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reflect.TypeOf(msg.GetMessage())).
		To(Equal(reflect.TypeOf(&protos.ClientToServer_RegisterReq{})))
	Expect(clientStream.State().String()).To(Equal(stream.Registering.String()))

	// finish registration
	err = clientServerConn.Send(&protos.ServerToClient{
		Message: &protos.ServerToClient_RegisterRes{
			RegisterRes: &protos.ClientRegisterRes{
				Status: &protos.Status{
					Type: protos.Status_OK,
				},
			},
		},
	})
	Expect(err).ShouldNot(HaveOccurred())

	// assert it has registered correctly
	test.AssertWithTimeout(
		func() bool { return clientStream.State() == stream.Registered }, condTimeout,
		func() { Expect(clientStream.State().String()).To(Equal(stream.Registered.String())) },
	)
}

func newDefaultStreamOpts() *stream.Options {
	return &stream.Options{
		ConnTimeout:        time.Duration(5) * time.Second,
		ResponseTimeout:    time.Duration(5) * time.Second,
		KeepAliveMaxPeriod: time.Duration(5) * time.Second,
		ServerReqsQueueLen: 10,
	}
}

var _ = Describe("Stream", func() {
	var (
		server          *mock.ControlPlaneServer
		recvServerReqCb func(*protos.ServerToClient) (bool, *protos.ClientToServer, error)
		clientStream    *stream.Stream[*protos.ServerToClient, *protos.ClientToServer]
	)

	BeforeEach(func() {
		var err error

		logger, err := logging.NewZapDev()
		Expect(err).ShouldNot(HaveOccurred())

		server = mock.NewControlPlaneServer(GinkgoT())

		handler := stream.NewClientHandler(func(m *protos.ServerToClient) (bool, *protos.ClientToServer, error) { return recvServerReqCb(m) })
		clientStream = stream.New(uuid.NewString(), newDefaultStreamOpts(), handler, logger)
	})

	AfterEach(func() {
		err := clientStream.Close(condTimeout)
		Expect(err).ShouldNot(HaveOccurred())

		server.Stop()
	})

	Describe("Connecting", func() {
		When("server is reachable", func() {
			It("should register", func() {
				end := make(chan struct{})
				server.SetClientHandlers(
					mock.RegisterClientHandler,
					func(stream protos.ControlPlane_ClientConnServer) error {
						end <- struct{}{}
						return nil
					},
				)
				server.Start(GinkgoT())

				err := clientStream.Connect(server.Addr())
				Expect(err).ShouldNot(HaveOccurred())
				<-end
			})

			It("should deregister when connection to server is lost", func() {
				registered := make(chan struct{})
				server.SetClientHandlers(
					mock.RegisterClientHandler,
					func(stream protos.ControlPlane_ClientConnServer) error {
						registered <- struct{}{}
						return nil
					},
				)
				server.Start(GinkgoT())

				err := clientStream.Connect(server.Addr())
				Expect(err).ShouldNot(HaveOccurred())
				<-registered

				server.Stop()
				test.AssertWithTimeout(
					func() bool { return clientStream.State() != stream.Registered }, condTimeout,
					func() { Expect(clientStream.State().String()).To(Not(Equal(stream.Registered.String()))) },
				)
			})

			It("should reregister when connection to server is lost and recovered", func() {
				registered := make(chan struct{})
				server.SetClientHandlers(
					mock.RegisterClientHandler,
					func(stream protos.ControlPlane_ClientConnServer) error {
						registered <- struct{}{}
						return nil
					},
				)
				server.Start(GinkgoT())

				err := clientStream.Connect(server.Addr())
				Expect(err).ShouldNot(HaveOccurred())
				<-registered

				server.Stop()
				test.AssertWithTimeout(
					func() bool { return clientStream.State() != stream.Registered }, condTimeout,
					func() { Expect(clientStream.State().String()).To(Not(Equal(stream.Registered.String()))) },
				)

				// make sure it did not reregister before restarting the server
				for len(registered) > 0 {
					<-registered
				}

				server.Start(GinkgoT())
				<-registered
			})

		})
	})

	Describe("Recveiving", func() {
		When("from server", func() {
			// TODO: Once we've got a real request from the server to the client to handle, replace req/res message types
			It("should forward the received msg to handler and send response back to server", func() {
				fakeServerReq := &protos.ServerToClient{
					Message: &protos.ServerToClient_SamplerStatsMsg{
						SamplerStatsMsg: &protos.ClientSamplerStatsMsg{
							SamplerStats: []*protos.ClientSamplerStats{
								{
									SamplerUid: "some_uid",
									SamplingStats: &protos.SamplerSamplingStats{
										SamplesEvaluated: uint64(rand.Int()),
										SamplesExported:  uint64(rand.Int()),
									},
								},
							},
						},
					},
				}

				fakeClientRes := &protos.ClientToServer{
					Message: &protos.ClientToServer_ListSamplersReq{},
				}

				registered := make(chan struct{})
				serverReceivedRes := make(chan struct{})
				server.SetClientHandlers(
					mock.RegisterClientHandler,
					func(stream protos.ControlPlane_ClientConnServer) error {
						registered <- struct{}{}
						return nil
					},
					func(stream protos.ControlPlane_ClientConnServer) error {
						Expect(stream.Send(fakeServerReq)).ShouldNot(HaveOccurred())

						fakeRes, err := stream.Recv()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(proto.Equal(fakeClientRes, fakeRes)).To(BeTrue())

						serverReceivedRes <- struct{}{}
						return nil
					},
				)
				server.Start(GinkgoT())

				clientReceivedReq := make(chan struct{})
				recvServerReqCb = func(msg *protos.ServerToClient) (bool, *protos.ClientToServer, error) {
					test.AssertWithTimeout(
						func() bool { return proto.Equal(fakeServerReq, msg) }, condTimeout,
						func() { Expect(proto.Equal(fakeServerReq, msg)).To(BeTrue()) },
					)

					clientReceivedReq <- struct{}{}
					return true, fakeClientRes, nil
				}

				err := clientStream.Connect(server.Addr())
				Expect(err).ShouldNot(HaveOccurred())

				<-registered
				<-clientReceivedReq
				<-serverReceivedRes
			})
		})
	})

	Describe("Sending", func() {
		When("to server", func() {
			It("should send request to server and receive its response", func() {
				listSamplersReq := &protos.ClientToServer{
					Message: &protos.ClientToServer_ListSamplersReq{
						ListSamplersReq: &protos.ClientListSamplersReq{},
					},
				}

				listSamplersRes := &protos.ServerToClient{
					Message: &protos.ServerToClient_ListSamplersRes{
						ListSamplersRes: &protos.ClientListSamplersRes{
							Status: &protos.Status{
								Type: protos.Status_OK,
							},
							Samplers: []*protos.Sampler{
								{
									Uid:  "some_uid",
									Name: "some_name",
									Config: &protos.SamplerConfig{
										SamplingRules: []*protos.SamplingRule{
											{
												Uid:      "some_sr_uid",
												Language: protos.SamplingRule_CEL,
												Rule:     "some_CEL_rule",
											},
										},
									},
								},
							},
						},
					},
				}

				registered := make(chan struct{})
				server.SetClientHandlers(
					mock.RegisterClientHandler,
					func(stream protos.ControlPlane_ClientConnServer) error {
						registered <- struct{}{}
						return nil
					},
					func(stream protos.ControlPlane_ClientConnServer) error {
						msg, err := stream.Recv()
						Expect(err).ShouldNot(HaveOccurred())
						Expect(proto.Equal(listSamplersReq, msg)).To(BeTrue())

						err = stream.Send(listSamplersRes)
						Expect(err).ShouldNot(HaveOccurred())
						return nil
					},
				)
				server.Start(GinkgoT())

				recvServerReqCb = func(_ *protos.ServerToClient) (bool, *protos.ClientToServer, error) {
					defer GinkgoRecover()
					return false, nil, nil
				}

				err := clientStream.Connect(server.Addr())
				Expect(err).ShouldNot(HaveOccurred())
				<-registered

				res, err := clientStream.SendReqToS(context.Background(), listSamplersReq)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(proto.Equal(listSamplersRes, res)).To(BeTrue())
			})
		})
	})

	Describe("Closing", func() {
		When("there are active connections", func() {
			It("should terminate them", func() {
				registered := make(chan struct{})
				closed := make(chan struct{})
				end := make(chan struct{})
				server.SetClientHandlers(
					mock.RegisterClientHandler,
					func(stream protos.ControlPlane_ClientConnServer) error {
						registered <- struct{}{}
						<-closed

						_, err := stream.Recv()
						Expect(err).Should(HaveOccurred())

						end <- struct{}{}
						return nil
					},
				)
				server.Start(GinkgoT())

				err := clientStream.Connect(server.Addr())
				Expect(err).ShouldNot(HaveOccurred())
				<-registered

				err = clientStream.Close(time.Microsecond)
				Expect(err).ShouldNot(HaveOccurred())
				closed <- struct{}{}

				<-end
			})
		})
	})
})

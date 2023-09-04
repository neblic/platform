package stream_test

//revive:disable:dot-imports
import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/controlplane/server/internal/defs"
	"github.com/neblic/platform/controlplane/server/internal/protocol/stream"
	"github.com/neblic/platform/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const condTimeout = time.Duration(1) * time.Second

func TestServerStream(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stream Suite")
}

type recvMsg struct {
	msg *protos.ClientToServer
	err error
}

type testClientStream struct {
	recvFromServerCb func(*protos.ServerToClient) error
	toServerMsgCh    chan recvMsg
}

func newTestClientStream() *testClientStream {
	return &testClientStream{
		toServerMsgCh: make(chan recvMsg),
	}
}

func (s *testClientStream) Send(msg *protos.ServerToClient) error {
	return s.recvFromServerCb(msg)
}

func (s *testClientStream) Recv() (*protos.ClientToServer, error) {
	recvMsg := <-s.toServerMsgCh
	return recvMsg.msg, recvMsg.err
}

func (s *testClientStream) sendToServer(msg *protos.ClientToServer, err error) {
	s.toServerMsgCh <- recvMsg{
		msg: msg,
		err: err,
	}
}

var _ = Describe("ServerStream", func() {
	var (
		clientStream *testClientStream
		serverStream *stream.Stream[*protos.ClientToServer, *protos.ServerToClient]

		recvServerReqCb stream.ClientRecvFromServerReqFunc
		stateChangeCb   stream.ClientStatusChangeFunc
	)

	setupTest := func() func() {
		// setup
		regResReceived := make(chan struct{}, 1)
		clientStream.recvFromServerCb = func(msg *protos.ServerToClient) error {
			defer func() {
				regResReceived <- struct{}{}
			}()

			Expect(reflect.TypeOf(msg.Message)).To(Equal(reflect.TypeOf(&protos.ServerToClient_RegisterRes{})))

			return nil
		}

		nextStatus := 0
		expectedStatus := []defs.Status{
			defs.RegisteredStatus,
			defs.UnregisteredStatus,
		}
		stateChanged := make(chan struct{}, 2)
		stateChangeCb = func(state defs.Status, uid control.ClientUID) error {
			defer func() {
				stateChanged <- struct{}{}
			}()

			Expect(state).To(Equal(expectedStatus[nextStatus]))
			nextStatus++

			return nil
		}

		// start
		streamClosed := make(chan struct{}, 1)
		go func() {
			defer GinkgoRecover()
			defer func() {
				streamClosed <- struct{}{}
			}()

			err := serverStream.Handle(clientStream)
			Expect(err).ShouldNot(HaveOccurred())
		}()

		// register
		clientStream.sendToServer(&protos.ClientToServer{
			Timestamp: timestamppb.New(time.Now()),
			ClientUid: "",
			Message: &protos.ClientToServer_RegisterReq{
				RegisterReq: &protos.ClientRegisterReq{},
			},
		}, nil)

		<-stateChanged // registered
		<-regResReceived

		return func() {
			err := serverStream.Close(condTimeout)
			Expect(err).ShouldNot(HaveOccurred())
			<-stateChanged // unregistered
			<-streamClosed
		}
	}

	BeforeEach(func() {
		clientStream = newTestClientStream()
		handler := stream.NewClientHandler(
			func(msg *protos.ClientToServer) (bool, *protos.ServerToClient, error) { return recvServerReqCb(msg) },
			func(state defs.Status, uid control.ClientUID) error { return stateChangeCb(state, uid) },
		)
		logger, err := logging.NewZapDev()
		Expect(err).To(Not(HaveOccurred()))

		serverStream = stream.New[*protos.ClientToServer, *protos.ServerToClient](
			logger,
			uuid.NewString(),
			&stream.Options{
				RegistrationReqTimeout: time.Second,
				ResponseTimeout:        time.Second,
				ReqsQueueLen:           10,
			},
			handler,
		)
	})

	AfterEach(func() {
		clientStream = nil
		serverStream = nil
	})

	Describe("Connecting", func() {
		When("server is reachable", func() {
			It("should register", func() {
				end := setupTest()

				end()
			})
		})
	})

	Describe("Receiving", func() {
		When("from client", func() {
			It("should forward the received msg to handler and send response back to client", func() {
				// * Client registers with server
				// * Client sends ListSamplersReq
				// * Server replies ListSamplersRes

				listSamplersRes := &protos.ClientListSamplersRes{
					Status: &protos.Status{
						Type: protos.Status_OK,
					},
					Samplers: []*protos.Sampler{
						{
							Uid:  "some_uid",
							Name: "some_name",
							Config: &protos.SamplerConfig{
								Streams: []*protos.Stream{
									{
										Uid: "some_stream_uid",
										Rule: &protos.Rule{
											Language:   protos.Rule_CEL,
											Expression: "some_CEL_rule",
										},
									},
								},
							},
						},
					},
				}

				end := setupTest()

				// check server response as received by client
				clientReceivedRes := make(chan struct{}, 1)
				clientStream.recvFromServerCb = func(msg *protos.ServerToClient) error {
					defer func() {
						clientReceivedRes <- struct{}{}
					}()

					Expect(reflect.TypeOf(msg.Message)).To(Equal(reflect.TypeOf(&protos.ServerToClient_ListSamplersRes{})))
					Expect(proto.Equal(listSamplersRes, msg.GetListSamplersRes())).To(BeTrue())

					return nil
				}

				// client sends request
				clientStream.sendToServer(&protos.ClientToServer{
					Message: &protos.ClientToServer_ListSamplersReq{
						ListSamplersReq: &protos.ClientListSamplersReq{},
					},
				}, nil)

				// server sends response to request
				serverReceivedReq := make(chan struct{}, 1)
				recvServerReqCb = func(msg *protos.ClientToServer) (bool, *protos.ServerToClient, error) {
					defer func() {
						serverReceivedReq <- struct{}{}
					}()

					Expect(reflect.TypeOf(msg.Message)).To(Equal(reflect.TypeOf(&protos.ClientToServer_ListSamplersReq{})))
					return true, &protos.ServerToClient{
						Message: &protos.ServerToClient_ListSamplersRes{
							ListSamplersRes: listSamplersRes,
						},
					}, nil
				}

				<-serverReceivedReq
				<-clientReceivedRes

				end()
			})
		})
	})

	Describe("Sending", func() {
		When("to client", func() {
			It("should send request to client and receive its response", func() {
				// * Client registers with server
				// * Server sends request (fake request)
				// * Client sends response (fake)

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

				end := setupTest()

				clientStream.recvFromServerCb = func(msg *protos.ServerToClient) error {
					Expect(proto.Equal(fakeServerReq, msg)).To(BeTrue())
					clientStream.sendToServer(fakeClientRes, nil)

					return nil
				}
				recvServerReqCb = func(*protos.ClientToServer) (bool, *protos.ServerToClient, error) {
					return false, nil, nil
				}

				res, err := serverStream.Send(fakeServerReq)
				Expect(err).To(Not(HaveOccurred()))
				Expect(proto.Equal(fakeClientRes, res)).To(BeTrue())

				end()
			})
		})
	})
})

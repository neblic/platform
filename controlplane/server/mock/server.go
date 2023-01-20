package mock

//revive:disable:dot-imports
import (
	"net"
	"reflect"

	"github.com/neblic/platform/controlplane/protos"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
)

type TestingT interface {
	Logf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type SamplerHandler func(stream protos.ControlPlane_SamplerConnServer) error
type ClientHandler func(stream protos.ControlPlane_ClientConnServer) error

func RegisterSamplerHandler(stream protos.ControlPlane_SamplerConnServer) error {
	// assert it has started registering
	msg, err := stream.Recv()
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reflect.TypeOf(msg.GetMessage())).
		To(Equal(reflect.TypeOf(&protos.SamplerToServer_RegisterReq{})))

	// finish registration
	err = stream.Send(&protos.ServerToSampler{
		Message: &protos.ServerToSampler_RegisterRes{
			RegisterRes: &protos.SamplerRegisterRes{
				Status: &protos.Status{
					Type: protos.Status_OK,
				},
			},
		},
	})
	Expect(err).ShouldNot(HaveOccurred())

	return nil
}

func RegisterClientHandler(stream protos.ControlPlane_ClientConnServer) error {
	// assert it has started registering
	msg, err := stream.Recv()
	Expect(err).ShouldNot(HaveOccurred())
	Expect(reflect.TypeOf(msg.GetMessage())).
		To(Equal(reflect.TypeOf(&protos.ClientToServer_RegisterReq{})))

	// finish registration
	err = stream.Send(&protos.ServerToClient{
		Message: &protos.ServerToClient_RegisterRes{
			RegisterRes: &protos.ClientRegisterRes{
				Status: &protos.Status{
					Type: protos.Status_OK,
				},
			},
		},
	})
	Expect(err).ShouldNot(HaveOccurred())

	return nil
}

type ControlPlaneServer struct {
	grpcServer *grpc.Server
	lis        net.Listener
	addr       net.Addr

	samplerHandlers []SamplerHandler
	clientHandlers  []ClientHandler

	protos.UnimplementedControlPlaneServer
}

func NewControlPlaneServer(t TestingT) *ControlPlaneServer {
	s := &ControlPlaneServer{}
	return s
}

func (s *ControlPlaneServer) SetSamplerHandlers(handlers ...SamplerHandler) {
	s.samplerHandlers = handlers
}

func (s *ControlPlaneServer) SetClientHandlers(handlers ...ClientHandler) {
	s.clientHandlers = handlers
}

func (s *ControlPlaneServer) SamplerConn(stream protos.ControlPlane_SamplerConnServer) error {
	defer GinkgoRecover()

	var err error
	for _, handler := range s.samplerHandlers {
		err = handler(stream)
		if err != nil {
			return err
		}
	}
	return err
}

func (s *ControlPlaneServer) ClientConn(stream protos.ControlPlane_ClientConnServer) error {
	defer GinkgoRecover()

	var err error
	for _, handler := range s.clientHandlers {
		err = handler(stream)
		if err != nil {
			return err
		}
	}
	return err
}

func (s *ControlPlaneServer) Start(t TestingT) {
	grpcServer := grpc.NewServer()
	protos.RegisterControlPlaneServer(grpcServer, s)
	s.grpcServer = grpcServer

	if s.addr == nil {
		lis, err := net.Listen("tcp", "localhost:")
		if err != nil {
			t.Errorf("error listening at 'localhost:': %s", err)
		}

		s.lis = lis
		s.addr = lis.Addr()
	} else {
		lis, err := net.Listen("tcp", s.addr.String())
		if err != nil {
			t.Errorf("error listening at 'localhost:': %s", err)
		}

		s.lis = lis
	}

	go func() {
		if err := s.grpcServer.Serve(s.lis); err != nil {
			t.Logf("grpc server exited with error: %s", err)
		}
	}()
}

func (s *ControlPlaneServer) Addr() string {
	return s.addr.String()
}

func (s *ControlPlaneServer) Stop() {
	s.grpcServer.Stop()
}

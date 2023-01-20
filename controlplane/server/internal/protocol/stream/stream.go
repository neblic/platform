package stream

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/logging"
	"google.golang.org/grpc/status"
)

type ToS interface {
	*protos.ClientToServer | *protos.SamplerToServer
}

type FromS interface {
	*protos.ServerToClient | *protos.ServerToSampler
}

type req[T ToS, F FromS] struct {
	fromSReq F
	toSRes   chan T
	timeout  *time.Timer
}

type Stream[T ToS, F FromS] struct {
	UID       string
	ServerUID string

	opts    *Options
	handler Handler[T, F]

	end           chan struct{}
	sendRequestCh chan *req[T, F]
	logger        logging.Logger
}

func New[T ToS, F FromS](logger logging.Logger, serverUID string, opts *Options, handler Handler[T, F]) *Stream[T, F] {
	if logger == nil {
		logger = logging.NewNopLogger()
	}

	return &Stream[T, F]{
		ServerUID: serverUID,

		opts:    opts,
		handler: handler,

		end:           make(chan struct{}),
		sendRequestCh: make(chan *req[T, F], opts.ReqsQueueLen),
		logger:        logger,
	}
}

func (s *Stream[ToS, FromS]) Handle(stream grpcStream[ToS, FromS]) error {
	recvCh := s.buildRecvCh(stream)

	if err := s.waitRegistrationReq(stream.Send, recvCh); err != nil {
		return err
	}
	defer s.deregister()

	resTimeout := make(chan struct{})
	reqs := make([]*req[ToS, FromS], 0)
loop:
	for {
		select {
		case <-s.end:
			s.logger.Debug("Received request to close stream")
			break loop
		case <-resTimeout:
			s.logger.Error("Time out while waiting for response")
			break loop
		case req := <-s.sendRequestCh:
			if err := stream.Send(req.fromSReq); err != nil {
				s.logger.Error(fmt.Sprintf("Error sending request: %s", err))
				break loop
			}

			req.timeout = time.AfterFunc(s.opts.ResponseTimeout, func() {
				resTimeout <- struct{}{}
			})

			reqs = append(reqs, req)
		case toSMsg, more := <-recvCh:
			if !more {
				break loop
			}

			if handled, res, err := s.handler.recvReqToS(toSMsg); err != nil {
				s.logger.Error(fmt.Sprintf("Error handling request: %s", err))
			} else {
				if handled {
					if res != nil {
						if err := stream.Send(res); err != nil {
							s.logger.Error(fmt.Sprintf("Error sending request response: %s", err))
						}
					}
				} else {
					if len(reqs) == 0 {
						s.logger.Error("Received response when there are no pending responses")
						break loop
					}
					reqs[0].timeout.Stop()
					reqs[0].toSRes <- toSMsg
					close(reqs[0].toSRes)

					reqs = reqs[1:]
				}
			}
		}
	}

	for _, req := range reqs {
		close(req.toSRes)
	}

	s.logger.Debug("Closing stream")

	return nil
}

func (s *Stream[FromS, ToS]) FromServerMsg() ToS {
	return s.handler.fromServerMsg(s.ServerUID)
}

func (s *Stream[ToS, FromS]) Send(fromS FromS) (ToS, error) {
	r := &req[ToS, FromS]{
		fromSReq: s.handler.fromServerMsg(s.UID),
		toSRes:   make(chan ToS),
	}
	r.fromSReq = fromS

	select {
	case s.sendRequestCh <- r:
	default:
		return nil, fmt.Errorf("can't process request, queue is full")
	}

	toSMsg, more := <-r.toSRes
	if !more {
		return nil, fmt.Errorf("response not received")
	}

	return toSMsg, nil
}

func (s *Stream[ToS, FromS]) Close(timeout time.Duration) error {
	timeoutTicker := time.NewTicker(timeout)
	select {
	case s.end <- struct{}{}:
	case <-timeoutTicker.C:
		return fmt.Errorf("timeout waiting for internal routines to process stop request")
	}

	return nil
}

func (s *Stream[ToS, FromS]) buildRecvCh(stream grpcStream[ToS, FromS]) chan ToS {
	recvCh := make(chan ToS)

	go func() {
	loop:
		for {
			clientToServerMsg, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					s.logger.Debug("Server disconnected the stream")
				} else {
					st, ok := status.FromError(err)
					if !ok {
						s.logger.Debug(fmt.Sprintf("Unknown error receiving message: %s", err))
					} else {
						s.logger.Debug(fmt.Sprintf("Error receiving message: %s", st))
					}
				}

				close(recvCh)
				break loop
			}

			recvCh <- clientToServerMsg
		}
	}()

	return recvCh
}

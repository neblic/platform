package stream

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/neblic/platform/controlplane/protos"
	"github.com/neblic/platform/logging"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

const minStreamDuration = time.Duration(60) * time.Second

type State int

const (
	Unknown State = iota
	Unregistered
	Registering
	Registered
)

func (s State) String() string {
	switch s {
	case Unknown:
		return "Unknown"
	case Unregistered:
		return "Unregistered"
	case Registering:
		return "Registering"
	case Registered:
		return "Registered"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

type FromS interface {
	*protos.ServerToClient | *protos.ServerToSampler
}

type ToS interface {
	*protos.ClientToServer | *protos.SamplerToServer
}

type req[F FromS, T ToS] struct {
	toSReq    T
	serverRes chan F
	timeout   *time.Timer
}

type fromSOrError[F FromS] struct {
	fromS F
	err   error
}

type Stream[F FromS, T ToS] struct {
	state     State
	uid       string
	serverUID string

	handler Handler[F, T]
	opts    *Options

	client               protos.ControlPlaneClient
	grpcConn             *grpc.ClientConn
	cancelCurrentContext func()

	errors      chan error
	stateChange chan State

	sendToSCh         chan *req[F, T]
	stopReconnections chan struct{}
	logger            logging.Logger
}

func New[F FromS, T ToS](uid string, opts *Options, h Handler[F, T], logger logging.Logger) *Stream[F, T] {
	s := &Stream[F, T]{
		uid: uid,

		handler: h,
		opts:    opts,

		sendToSCh:         make(chan *req[F, T], opts.ServerReqsQueueLen),
		stopReconnections: make(chan struct{}, 1),
		logger:            logger,
	}

	return s
}

func (s *Stream[FromS, ToS]) Connect(serverAddr string) error {
	if s.grpcConn != nil {
		return fmt.Errorf("already connected")
	}

	s.logger.Debug(fmt.Sprintf("Connecting to %s, options: %+v", serverAddr, s.opts))

	var dialOpts []grpc.DialOption

	// TLS configuration
	var transportCredentials credentials.TransportCredentials
	if s.opts.TLS.Enable {
		if s.opts.TLS.CACertPath == "" {
			systemCrts, err := x509.SystemCertPool()
			if err != nil {
				return fmt.Errorf("error getting system root certificate pool: %w", err)
			}

			transportCredentials = credentials.NewTLS(&tls.Config{
				RootCAs: systemCrts,
			})
		} else {
			var err error
			transportCredentials, err = credentials.NewClientTLSFromFile(s.opts.TLS.CACertPath, "")
			if err != nil {
				return fmt.Errorf("error loading TLS CA: %w", err)
			}
		}
	} else {
		transportCredentials = insecure.NewCredentials()
	}
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(transportCredentials))

	// Auth configuration
	switch s.opts.Auth.Type {
	case "bearer":
		t := &oauth2.Token{
			AccessToken: s.opts.Auth.Bearer.Token,
			TokenType:   "Bearer",
		}

		dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(oauth.NewOauthAccess(t)))
	case "":
		// nothing to do
	default:
		return fmt.Errorf("invalid auth type %s", s.opts.Auth.Type)
	}

	// Connection configuration
	ctx := context.Background()
	if s.opts.Block {
		dialOpts = append(dialOpts, grpc.WithBlock())

		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), s.opts.ConnTimeout)
		defer cancel()
	}

	dialOpts = append(dialOpts,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time: s.opts.KeepAliveMaxPeriod,
		}),
	)

	grpcConn, err := grpc.DialContext(ctx, serverAddr, dialOpts...)
	if err != nil {
		return fmt.Errorf("error connecting to addr %s: %w", serverAddr, err)
	}

	s.logger.Debug(fmt.Sprintf("Connected to %s", serverAddr))
	s.client = protos.NewControlPlaneClient(grpcConn)
	s.grpcConn = grpcConn

	go s.reconnectRoutine()

	return nil
}

func (s *Stream[FromS, ToS]) ToServerMsg() ToS {
	return s.handler.toServerMsg(s.serverUID)
}

func (s *Stream[FromS, ToS]) SendToS(ctx context.Context, toS ToS) error {
	if s.state != Registered {
		return fmt.Errorf("can't send message to server, stream not registered")
	}

	r := &req[FromS, ToS]{
		toSReq: toS,
	}

	select {
	case s.sendToSCh <- r:
	default:
		return fmt.Errorf("can't send message to server, queue is full")
	}

	return nil
}

func (s *Stream[FromS, ToS]) SendReqToS(ctx context.Context, toS ToS) (FromS, error) {
	r := &req[FromS, ToS]{
		serverRes: make(chan FromS),
		toSReq:    toS,
	}

	select {
	case s.sendToSCh <- r:
	default:
		return nil, fmt.Errorf("can't send message to server, queue is full")
	}

	select {
	case fromSMsg, more := <-r.serverRes:
		if !more {
			return nil, fmt.Errorf("response not received from server")
		}
		return fromSMsg, nil
	case <-ctx.Done():
		// TODO: Given that we can't just discard the response once arrives, for now, we force a reconnection
		s.cancelCurrentContext()

		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}
}

func (s *Stream[FromS, ToS]) State() State {
	return s.state
}

func (s *Stream[FromS, ToS]) StateChanges() chan State {
	if s.stateChange == nil {
		s.stateChange = make(chan State)
	}

	return s.stateChange
}

func (s *Stream[FromS, ToS]) Errors() chan error {
	if s.errors == nil {
		s.errors = make(chan error)
	}

	return s.errors
}

func (s *Stream[FromS, ToS]) ServerUID() string {
	return s.serverUID
}

func (s *Stream[FromS, ToS]) Close(timeout time.Duration) error {
	if s.grpcConn == nil {
		return nil
	}
	s.logger.Debug("Closing stream")

	timeoutTicker := time.NewTicker(timeout)
	select {
	case s.stopReconnections <- struct{}{}:
	case <-timeoutTicker.C:
		s.logger.Error("Timeout waiting for internal routines to process stop request")
	}

	var closeErr error
	if err := s.grpcConn.Close(); err != nil {
		closeErr = fmt.Errorf("error closing connection: %w", err)
	}

	s.grpcConn = nil

	return closeErr
}

func (s *Stream[FromS, ToS]) setState(st State) {
	s.state = st

	if s.stateChange != nil {
		s.stateChange <- st
	}
}

func (s *Stream[FromS, ToS]) notifyError(err error) {
	s.logger.Debug(err.Error())

	if s.errors != nil {
		s.errors <- err
	}
}

func initBackOff() backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 0
	b.Reset()

	return b
}

func (s *Stream[FromS, ToS]) reconnectRoutine() {
	s.logger.Debug("Started reconnection routine")

	backOff := initBackOff()

	initialConnection := true
loop:
	for {
		select {
		case <-s.stopReconnections:
			s.logger.Debug("Received request to stop reconnecting")

			break loop
		default:
			if !initialConnection {
				backoffDuration := backOff.NextBackOff()
				s.logger.Debug(fmt.Sprintf("Reconnecting in %s", backoffDuration))

				sleepTicker := time.NewTicker(backoffDuration)
				select {
				case <-sleepTicker.C:
				case <-s.stopReconnections:
					break loop
				}
			} else {
				initialConnection = false
			}

			// Needed to free resources associated with the previous context
			if s.cancelCurrentContext != nil {
				s.cancelCurrentContext()
				s.cancelCurrentContext = nil
			}

			ctx, cancel := context.WithCancel(context.Background())
			s.cancelCurrentContext = cancel

			stream, err := s.getNewServerStream(ctx)
			if err != nil {
				s.notifyError(
					fmt.Errorf("error getting server stream: %w", err),
				)

				continue
			}

			streamStart := time.Now()
			if err := s.handleStream(stream); err != nil {
				s.notifyError(
					fmt.Errorf("error handling stream: %w", err),
				)

				streamEnd := time.Now()
				streamDuration := streamEnd.Sub(streamStart)
				if streamDuration > minStreamDuration {
					backOff.Reset()
				}

				continue
			}

			backOff.Reset()
		}
	}
}

func (s *Stream[FromS, ToS]) getNewServerStream(ctx context.Context) (grpcStream[FromS, ToS], error) {
	stream, err := s.handler.grpcStream(ctx, s.client)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, fmt.Errorf("%w: couldn't connect to server: %s", ErrRegistrationFailure, err)
		}
		return nil, fmt.Errorf("%w: couldn't connect to server: %s", ErrRegistrationFailure, st.Message())
	}

	return stream, nil
}

func (s *Stream[FromS, ToS]) getRecvCh(stream grpcStream[FromS, ToS]) chan fromSOrError[FromS] {
	recvCh := make(chan fromSOrError[FromS])

	go func() {
	loop:
		for {
			fromS, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					err = fmt.Errorf("server disconnected the stream")
				} else {
					st, ok := status.FromError(err)
					if !ok {
						err = fmt.Errorf("unknown error received from server: %w", err)
					} else {
						err = fmt.Errorf("error received from server: %s", st.Message())
					}
				}

				s.notifyError(err)

				recvCh <- fromSOrError[FromS]{
					err: err,
				}
				close(recvCh)
				break loop
			}

			recvCh <- fromSOrError[FromS]{
				fromS: fromS,
			}
		}
	}()

	return recvCh
}

// handleStream return nil error if the connection gets closed after successfully registered
func (s *Stream[FromS, ToS]) handleStream(stream grpcStream[FromS, ToS]) error {
	recvCh := s.getRecvCh(stream)

	if err := s.register(s.uid, stream.Send, recvCh, s.opts.ResponseTimeout); err != nil {
		return fmt.Errorf("%w: %s", ErrRegistrationFailure, err)
	}
	defer s.deregister()

	var err error
	resTimeout := make(chan struct{})
	reqs := make([]*req[FromS, ToS], 0)
loop:
	for {
		select {
		case <-resTimeout:
			err = fmt.Errorf("time out while waiting for server response")
			break loop
		case req := <-s.sendToSCh:
			// s.logger.Debug("Sending request to server")

			if streamErr := stream.Send(req.toSReq); err != nil {
				err = fmt.Errorf("error sending request: %s", streamErr)
				break loop
			}

			// if there is no response channel it is assumed the server will not send a response
			if req.serverRes != nil {
				req.timeout = time.AfterFunc(s.opts.ResponseTimeout, func() {
					resTimeout <- struct{}{}
				})

				reqs = append(reqs, req)
			}
		case fromSMsgOrErr, more := <-recvCh:
			if !more || fromSMsgOrErr.err != nil {
				err = fromSMsgOrErr.err
				break loop
			}

			if handled, res, err := s.handler.recvServerReq(fromSMsgOrErr.fromS); err != nil {
				s.notifyError(
					fmt.Errorf("error handling server request: %w", err),
				)
			} else {
				if handled {
					if res != nil {
						if err := stream.Send(res); err != nil {
							s.notifyError(
								fmt.Errorf("error sending request response: %w", err),
							)
						}
					}
				} else {
					if len(reqs) == 0 {
						err = errors.New("Received response when there are no pending responses")
						break loop
					}
					// s.logger.Debug("Received server response")

					reqs[0].timeout.Stop()
					reqs[0].serverRes <- fromSMsgOrErr.fromS
					close(reqs[0].serverRes)

					reqs = reqs[1:]
				}
			}
		}
	}

	for _, req := range reqs {
		close(req.serverRes)
	}

	return fmt.Errorf("%w: %s", ErrConnectionFailure, err)
}

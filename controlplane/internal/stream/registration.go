package stream

import (
	"fmt"
	"time"
)

func (s *Stream[FromS, ToS]) register(uid string, sendMsg func(ToS) error, recvCh chan fromSOrError[FromS], timeout time.Duration) error {
	if err := s.startRegistration(uid, sendMsg); err != nil {
		return err
	}

	registrationTimeout := time.NewTimer(timeout)
	select {
	case <-registrationTimeout.C:
		return fmt.Errorf("timed out while waiting for register response")
	case serverToClientMsgOrErr, ok := <-recvCh:
		if !ok {
			return fmt.Errorf("connection closed during registration")
		}

		if serverToClientMsgOrErr.err != nil {
			return serverToClientMsgOrErr.err
		}

		if err := s.finishRegistration(serverToClientMsgOrErr.fromS); err != nil {
			return err
		}

		return nil
	}
}

func (s *Stream[FromS, ToS]) startRegistration(uid string, sendMsg func(ToS) error) error {
	s.logger.Debug("Starting registration")

	toServerMsg := s.handler.regReqMsg(uid)

	if err := sendMsg(toServerMsg); err != nil {
		return fmt.Errorf("error sending registration request: %w", err)
	}

	s.setState(Registering)

	return nil
}

func (s *Stream[FromS, ToS]) finishRegistration(fromSMsg FromS) error {
	var (
		serverUID string
		err       error
	)

	if serverUID, err = s.handler.validateRegResMsg(fromSMsg); err != nil {
		return err
	}

	if s.state != Registering {
		return fmt.Errorf("received unexpected registration response while not being actively registering")
	}

	s.logger.Debug("Received registration response, finishing registration")

	s.serverUID = serverUID
	s.setState(Registered)

	return nil
}

func (s *Stream[FromS, ToS]) deregister() {
	s.logger.Debug("Deregistering")

	s.serverUID = ""
	s.setState(Unregistered)
}

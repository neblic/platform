package stream

import (
	"fmt"
	"time"
)

func (s *Stream[ToS, FromS]) waitRegistrationReq(sendMsg func(FromS) error, recvCh chan ToS) error {
	reqTimeout := time.NewTimer(s.opts.RegistrationReqTimeout)

	select {
	case <-reqTimeout.C:
		return fmt.Errorf("timed out while waiting for register request")
	case clientToServerMsg, ok := <-recvCh:
		if !ok {
			return fmt.Errorf("connection closed during registration")
		}

		uid, regRes, err := s.handler.startRegistration(clientToServerMsg)
		if err != nil {
			return err
		}

		if err := sendMsg(regRes); err != nil {
			return fmt.Errorf("error sending register request response: %w", err)
		}

		s.UID = uid
		s.logger = s.logger.With("stream_uid", s.UID)

		if err := s.handler.finishRegistration(clientToServerMsg); err != nil {
			return err
		}

		return nil
	}
}

func (s *Stream[ToS, FromS]) deregister() {
	if err := s.handler.deregister(s.UID); err != nil {
		s.logger.Error(fmt.Sprintf("error deregistering: %s", err))
	}

	s.UID = ""
}

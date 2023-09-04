package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/neblic/platform/controlplane/control"
	"github.com/neblic/platform/controlplane/server/internal/defs"
	"github.com/neblic/platform/logging"
)

var (
	ErrUnkownClient = errors.New("unknown client")
)

type ClientRegistry struct {
	clients map[control.ClientUID]*defs.Client

	logger logging.Logger
	m      sync.Mutex
}

func NewClientRegistry(logger logging.Logger) (*ClientRegistry, error) {
	if logger == nil {
		logger = logging.NewNopLogger()
	}

	return &ClientRegistry{
		clients: make(map[control.ClientUID]*defs.Client),
		logger:  logger,
	}, nil
}

func (cr *ClientRegistry) getClient(uid control.ClientUID) (*defs.Client, error) {
	foundClient, ok := cr.clients[uid]
	if !ok {
		return nil, fmt.Errorf("%w, uid: %s", ErrUnkownClient, uid)
	}

	return foundClient, nil
}

func (cr *ClientRegistry) Register(uid control.ClientUID) error {
	cr.m.Lock()
	defer cr.m.Unlock()

	knownClient, err := cr.getClient(uid)
	if err != nil && !errors.Is(err, ErrUnkownClient) {
		return err
	} else if err == nil {
		if knownClient.Status == defs.RegisteredStatus {
			cr.logger.Error("reregistering an already registered client", "client_id", uid)
		}
	}

	client := &defs.Client{
		UID:    uid,
		Status: defs.RegisteredStatus,
	}
	cr.clients[uid] = client

	return nil
}

func (cr *ClientRegistry) Deregister(UID control.ClientUID) error {
	cr.m.Lock()
	defer cr.m.Unlock()

	_, err := cr.getClient(UID)
	if errors.Is(err, ErrUnkownClient) {
		cr.logger.Debug("deregistering unknown client, nothing to do", "client_uid", UID)

		return nil
	} else if err != nil {
		return err
	}

	delete(cr.clients, UID)

	return nil
}

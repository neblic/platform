package stream

import (
	"encoding/json"
	"fmt"
	"time"
)

type TLSOptions struct {
	// Enable determines if the server implements TLS. If TLSCert is not set,
	// it t will use the system CA certificates to validate the server certificate.
	// By default, it is disabled.
	Enable bool
	// CACertPath defines te path to the certificate CA used to validate the
	// server certificate.
	CACertPath string
}

type BearerOptions struct {
	// Bearer token
	Token string
}

type AuthOptions struct {
	// Type specifies the authentication type, supported values: 'bearer'
	// If empty, no authentication enabled.
	Type string
	// Bearer configures the bearer authentication type..
	Bearer BearerOptions
}

type Options struct {
	// Block determines if the initial connection request will block until it
	// successfully connects.
	Block bool
	// ConnTimeout is a duration for the maximum amount of time to wait
	// during the initial connection request. Only applicable if Block is true.
	ConnTimeout time.Duration
	// ResponseTimeout is a duration for the maximum amount of time to wait
	// for a response to a request send to the server.
	ResponseTimeout time.Duration
	// KeepAliveMaxPeriod is a duration that defines the maximum period to
	// send keep alive messages. Minimum value is 10s. A larger period reduces
	// network load but increases the time required to detect when the stream
	// disconnects.
	KeepAliveMaxPeriod time.Duration
	// ServerReqsQueueLen defines how many server requests are allowed to be queued.
	ServerReqsQueueLen int
	// TLS related options
	TLS TLSOptions
	// Auth related options
	Auth AuthOptions
}

func NewOptionsDefault() *Options {
	return &Options{
		Block:              false,
		ConnTimeout:        time.Second * time.Duration(30),
		ResponseTimeout:    time.Second * time.Duration(10),
		KeepAliveMaxPeriod: time.Second * time.Duration(10),
		ServerReqsQueueLen: 10,
		TLS: TLSOptions{
			Enable:     false,
			CACertPath: "",
		},
		Auth: AuthOptions{
			Type: "",
		},
	}
}

func (o Options) String() string {
	// filter out sensitive fields
	o.Auth.Bearer.Token = "REDACTED"

	jsonOpt, _ := json.Marshal(o)
	return fmt.Sprintf("%s", jsonOpt)
}

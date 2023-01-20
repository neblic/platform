package stream

import "time"

type Options struct {
	// RegistrationReqTimeout is a duration for the maximum amount of time to wait
	// for a registration request after a new client/sampler connection.
	RegistrationReqTimeout time.Duration
	// ResponseTimeout is a duration for the maximum amount of time to wait
	// for a response to a request send to the client/sampler.
	ResponseTimeout time.Duration
	// ServerReqsQueueLen defines how many client/sampler requests are allowed to be queued.
	ReqsQueueLen int
}

func NewOptionsDefault() *Options {
	return &Options{
		RegistrationReqTimeout: time.Duration(10) * time.Second,
		ResponseTimeout:        time.Duration(10) * time.Second,
		ReqsQueueLen:           10,
	}
}

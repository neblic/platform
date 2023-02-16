package mock

import (
	"context"
	"net"
	"sync"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Copyright The OpenTelemetry Authors
// Licensed under the Apache License, Version 2.0 (the "License");
type mockReceiver struct {
	Srv          *grpc.Server
	RequestCount *atomic.Int32
	TotalItems   *atomic.Int32
	mux          sync.Mutex
	metadata     metadata.MD
	exportError  error
}

func (r *mockReceiver) getMetadata() metadata.MD {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.metadata
}

func (r *mockReceiver) setExportError(err error) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.exportError = err
}

type LogsReceiver struct {
	mockReceiver
	lastRequest plog.Logs

	plogotlp.UnimplementedGRPCServer
}

func (r *LogsReceiver) Export(ctx context.Context, req plogotlp.ExportRequest) (plogotlp.ExportResponse, error) {
	r.RequestCount.Inc()
	ld := req.Logs()
	r.TotalItems.Add(int32(ld.LogRecordCount()))
	r.mux.Lock()
	defer r.mux.Unlock()
	r.lastRequest = ld
	r.metadata, _ = metadata.FromIncomingContext(ctx)
	return plogotlp.NewExportResponse(), r.exportError
}

func (r *LogsReceiver) GetLastRequest() plog.Logs {
	r.mux.Lock()
	defer r.mux.Unlock()
	return r.lastRequest
}

func OtlpLogsReceiverOnGRPCServer(ln net.Listener) *LogsReceiver {
	rcv := &LogsReceiver{
		mockReceiver: mockReceiver{
			Srv:          grpc.NewServer(),
			RequestCount: atomic.NewInt32(0),
			TotalItems:   atomic.NewInt32(0),
		},
	}

	// Now run it as a gRPC server
	plogotlp.RegisterGRPCServer(rcv.Srv, rcv)
	go func() {
		_ = rcv.Srv.Serve(ln)
	}()

	return rcv
}

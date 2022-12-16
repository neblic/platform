package docs_test

import (
	"testing"

	// --8<-- [start:InterceptorInitImport]
	"context"

	neblicgrpc "github.com/neblic/platform/sampler/instrumentation/google.golang.org/grpc"
	"google.golang.org/grpc"
	// --8<-- [end:InterceptorInitImport]
)

// --8<-- [start:InterceptorInit]
func initInterceptor() *grpc.ClientConn {
	interceptor := neblicgrpc.UnaryClientInterceptor()
	conn, _ := grpc.DialContext(context.Background(), "server-addr:port",
		grpc.WithUnaryInterceptor(interceptor))

	return conn
}

// --8<-- [end:InterceptorInit]

func TestInitInetrceptor(t *testing.T) {
	initInterceptor()
}

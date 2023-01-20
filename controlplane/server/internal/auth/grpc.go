package auth

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func NewAuthBearerStreamInterceptor(token string) grpc.StreamServerInterceptor {
	bearerString := fmt.Sprintf("Bearer %s", token)

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, found := metadata.FromIncomingContext(ss.Context())
		if !found {
			return grpc.Errorf(codes.Unauthenticated, "no metadata in context")
		}

		bearerAuth := md.Get("authorization")
		if len(bearerAuth) == 0 {
			return grpc.Errorf(codes.Unauthenticated, "no authorization found")
		}

		if bearerAuth[0] != bearerString {
			return grpc.Errorf(codes.Unauthenticated, "invalid token")
		}

		return handler(srv, ss)
	}
}

package bearerauthextension

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension/auth"
)

var (
	errNoAuth             = errors.New("no auth provided")
	errInvalidCredentials = errors.New("invalid credentials")
)

type bearerAuth struct {
	token string

	bearerString string
}

func (ba *bearerAuth) serverStart(ctx context.Context, host component.Host) error {
	ba.bearerString = fmt.Sprintf("Bearer %s", ba.token)

	return nil
}

func (ba *bearerAuth) authenticate(ctx context.Context, headers map[string][]string) (context.Context, error) {
	auth := getAuthHeader(headers)
	if auth == "" {
		return ctx, errNoAuth
	}

	if auth != ba.bearerString {
		return ctx, errInvalidCredentials
	}

	return ctx, nil
}

func getAuthHeader(h map[string][]string) string {
	const metadataKey = "authorization"

	authHeaders, ok := h[metadataKey]
	if !ok {
		return ""
	}

	if len(authHeaders) == 0 {
		return ""
	}

	return authHeaders[0]
}

func newServerAuthExtension(cfg *Config) (auth.Server, error) {
	ba := bearerAuth{
		token: cfg.Token,
	}

	return auth.NewServer(
		auth.WithServerStart(ba.serverStart),
		auth.WithServerAuthenticate(ba.authenticate),
	), nil
}

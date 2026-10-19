package controller

import (
	"context"
	"errors"

	"github.com/rfizzle/go-starter/internal/api/v1"
)

// ErrInvalidCredentials indicates the presented email/password combination is
// invalid. Authenticator implementations return it (or wrap it) so Login can
// map it to a bounded, uninformative response without this package importing
// a concrete credential module.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Authenticator verifies an email/password pair and returns the subject to
// embed in an issued token — the user's email, matching the convention
// GetCurrentUser already uses when resolving a token's subject to a stored
// user. It is satisfied by a small composition-root adapter over
// internal/modules/credential.Service; this package never imports that module,
// because the controller only maps transport and the composition root adapts.
// See docs/decisions.md#d1-module-and-platform-boundaries
type Authenticator interface {
	Authenticate(ctx context.Context, email, password string) (string, error)
}

// Login verifies the request's credentials and, on success, issues a signed
// bearer token for the resulting subject. An unconfigured Authenticator or
// Tokens generator, an unknown email, and a wrong password all return the
// same LoginUnauthorized so a client cannot distinguish "no such account"
// from "wrong password" from "login is disabled." Any other Authenticate or
// token-signing error is returned as-is (mapped to a 500 by ogen's default
// error handling, matching GetCurrentUser's precedent for unexpected errors).
func (c *Controller) Login(ctx context.Context, req *api.LoginRequest) (api.LoginRes, error) {
	if c.options.Authenticator == nil || c.options.Tokens == nil {
		return &api.LoginUnauthorized{}, nil
	}

	subject, err := c.options.Authenticator.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return &api.LoginUnauthorized{}, nil
		}
		return nil, err
	}

	token, err := c.options.Tokens.NewToken(subject)
	if err != nil {
		return nil, err
	}

	return &api.LoginResponse{Token: token}, nil
}

package controller

import (
	"context"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/auth"
)

// compile-time check that Controller satisfies ogen's SecurityHandler.
var _ api.SecurityHandler = (*Controller)(nil)

// HandleBearerAuth authenticates the bearer token ogen extracted from the
// Authorization header. On success it returns a context carrying the token's
// subject, which flows through to the operation handler (e.g. GetCurrentUser).
// A non-nil error causes ogen to reject the request before the handler runs.
func (c *Controller) HandleBearerAuth(ctx context.Context, _ api.OperationName, t api.BearerAuth) (context.Context, error) {
	// Fail closed: a protected operation with no configured validator must not
	// be treated as authenticated.
	if c.options.Validator == nil {
		return ctx, ogenerrors.ErrSecurityRequirementIsNotSatisfied
	}

	subject, err := c.options.Validator.ValidateToken(t.Token)
	if err != nil {
		return ctx, err
	}

	return auth.WithSubject(ctx, subjectContextKey, subject)
}

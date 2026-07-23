package controller

import (
	"context"
	"fmt"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/auth"
	"github.com/rfizzle/go-starter/internal/domain"
)

// subjectContextKey is the key under which the authenticated subject is stored
// in the request context by HandleBearerAuth and read back by GetCurrentUser.
const subjectContextKey = "subject"

type Controller struct {
	api.Handler
	options *Options
}

// UserReader looks up users for request handlers. It is the read subset of the
// user domain the controller needs, kept as an interface so handlers can be
// tested without a database. *user.Repository satisfies it without coupling
// this transport edge to the concrete module.
type UserReader interface {
	ByEmail(ctx context.Context, email string) (domain.User, error)
}

// ReadinessChecker evaluates whether the application's lifecycle and required
// dependencies can currently serve traffic.
type ReadinessChecker interface {
	Ready(ctx context.Context) bool
}

type Options struct {
	Readiness ReadinessChecker
	// Users resolves the authenticated subject to a stored user record. It may
	// be nil, in which case protected handlers fall back to the token subject
	// alone without a database lookup.
	Users UserReader
	// Validator authenticates bearer tokens for protected operations. It may be
	// nil when the API exposes no protected operations; requests to protected
	// operations then fail closed with an unauthorized error.
	Validator auth.TokenValidator
}

func NewController(options *Options) (*Controller, error) {
	// Validate options
	if options == nil {
		return nil, fmt.Errorf("options is required")
	}

	return &Controller{
		options: options,
	}, nil
}

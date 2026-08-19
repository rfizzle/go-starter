// Package controller maps the generated OpenAPI transport in internal/api/v1
// onto business behavior. It is a quarantined edge: it decodes requests,
// resolves the authenticated subject, and translates business errors into the
// bounded responses the specification defines.
//
// It owns no business logic. Its collaborators — user lookup, credential
// authentication, token validation and issuance, readiness — are small
// consumer-owned interfaces declared here and satisfied by concrete types
// constructed in internal/app, so this package never imports a business
// module.
package controller

import (
	"context"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/auth"
	"github.com/rfizzle/go-starter/internal/domain"
)

// subjectContextKey is the key under which the authenticated subject is stored
// in the request context by HandleBearerAuth and read back by GetCurrentUser.
const subjectContextKey = "subject"

type Controller struct {
	api.Handler
	options Options
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
	// Authenticator verifies login credentials and returns the subject to embed
	// in an issued token. It may be nil, in which case Login fails closed with
	// the same bounded response as invalid credentials.
	Authenticator Authenticator
	// Tokens issues signed bearer tokens for successful logins. It may be nil
	// when no JWT secret is configured, in which case Login fails closed. It is
	// the same auth.Generator implementation (currently *auth.JWT) used to
	// build Validator, so tokens Login issues validate against this server's
	// own /me.
	Tokens auth.Generator
}

// NewController builds the transport controller over the given collaborators.
// Every Options field may be nil; the handlers fail closed as documented on
// each field, so construction itself cannot fail.
func NewController(options Options) *Controller {
	return &Controller{options: options}
}

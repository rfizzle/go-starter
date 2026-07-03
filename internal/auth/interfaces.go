package auth

import (
	"net/http"

	"go.uber.org/zap"
)

// Generator interface for token and cookie generation
type Generator interface {
	NewToken(subject string) (string, error)
	NewCookie(subject string) (*http.Cookie, error)
}

// Validator interface for request validation
type Validator interface {
	Validate(r *http.Request) (string, error)
}

// TokenValidator validates an already-extracted token string and returns its
// subject. Transport layers that isolate the token before dispatch (e.g. ogen's
// SecurityHandler, which receives the bearer token directly) depend on this
// rather than the request-based Validator.
type TokenValidator interface {
	ValidateToken(token string) (string, error)
}

// Service interface defines the core authentication operations
type Service interface {
	// Middleware returns HTTP middleware for authentication
	Middleware() func(next http.Handler) http.Handler

	// RequireAuthentication returns middleware that enforces authentication
	RequireAuthentication() func(next http.Handler) http.Handler

	// RequireAuthenticationWithCustomResponse allows custom error handling
	RequireAuthenticationWithCustomResponse(errorHandler func(w http.ResponseWriter, r *http.Request, err error)) func(next http.Handler) http.Handler

	// AuditMiddleware returns middleware for audit logging
	AuditMiddleware(auditor AuditLogger) func(next http.Handler) http.Handler

	// Logger returns the service's logger
	Logger() *zap.Logger
}

// Provider interface for authentication providers (for backward compatibility)
type Provider interface {
	Service
}

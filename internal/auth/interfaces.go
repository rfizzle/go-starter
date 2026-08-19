package auth

import (
	"net/http"
)

// Generator interface for token generation
type Generator interface {
	NewToken(subject string) (string, error)
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

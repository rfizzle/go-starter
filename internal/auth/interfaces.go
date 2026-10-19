package auth

// Generator mints a signed token for a subject.
type Generator interface {
	NewToken(subject string) (string, error)
}

// TokenValidator validates an already-extracted token string and returns its
// subject. Transport layers that isolate the token before dispatch (e.g. ogen's
// SecurityHandler, which receives the bearer token directly) depend on this.
type TokenValidator interface {
	ValidateToken(token string) (string, error)
}

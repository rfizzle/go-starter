// Package auth provides token validation, token generation, and the
// request-context carrier for the authenticated subject.
//
// The package is a leaf consumed by the generated-transport edge: ogen's
// SecurityHandler validates a bearer token through a TokenValidator and stores
// the resulting subject with WithSubject; handlers read it back with
// GetSubjectFromContext. JWT is the bundled HMAC (HS256) implementation of both
// Validator and Generator.
//
// Example:
//
//	validator, err := auth.NewJWT([]byte(secret), auth.WithIssuer("go-starter"))
package auth

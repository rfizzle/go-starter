package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSigningMethod is the signing method used for HMAC (HS256) tokens. Keeping
// it fixed prevents algorithm-confusion attacks where an attacker swaps the
// header's "alg" for one the server would validate differently (e.g. "none").
var jwtSigningMethod = jwt.SigningMethodHS256

// JWT is a concrete HMAC (HS256) implementation of both the TokenValidator and
// Generator interfaces. It signs and verifies tokens with a shared secret,
// making it the simplest option for a service that issues its own tokens.
type JWT struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration

	now func() time.Time
}

// compile-time checks that JWT satisfies both interfaces.
var (
	_ TokenValidator = (*JWT)(nil)
	_ Generator      = (*JWT)(nil)
)

// JWTOption configures a JWT instance.
type JWTOption func(*JWT)

// WithIssuer sets the issuer ("iss") claim written on generated tokens and
// required on validated tokens. When empty (default), the issuer is not checked.
func WithIssuer(issuer string) JWTOption {
	return func(j *JWT) { j.issuer = issuer }
}

// WithAudience sets the audience ("aud") claim written on generated tokens and
// required on validated tokens. When empty (default), the audience is not checked.
func WithAudience(audience string) JWTOption {
	return func(j *JWT) { j.audience = audience }
}

// WithTTL sets the lifetime of generated tokens. Defaults to one hour.
func WithTTL(ttl time.Duration) JWTOption {
	return func(j *JWT) { j.ttl = ttl }
}

// withClock overrides the time source; used by tests to make expiry deterministic.
func withClock(now func() time.Time) JWTOption {
	return func(j *JWT) { j.now = now }
}

// NewJWT builds a JWT validator/generator from a shared HMAC secret. The secret
// must be non-empty; a service booted without one would accept forged tokens.
func NewJWT(secret []byte, options ...JWTOption) (*JWT, error) {
	if len(secret) == 0 {
		return nil, fmt.Errorf("jwt secret is required")
	}

	j := &JWT{
		secret: secret,
		ttl:    time.Hour,
		now:    time.Now,
	}

	for _, option := range options {
		option(j)
	}

	return j, nil
}

// ValidateToken verifies a raw token string and returns its subject. It is the
// entry point used by transport layers (e.g. ogen's SecurityHandler) that have
// already isolated the token from the request.
func (j *JWT) ValidateToken(tokenString string) (string, error) {
	if tokenString == "" {
		return "", &Error{Err: ErrUnauthorized}
	}

	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwtSigningMethod.Alg()}),
		jwt.WithTimeFunc(j.now),
		// Require an "exp" claim. Without this, a token forged with the shared
		// secret but no expiry would be accepted as a permanent credential —
		// fail-open. Tokens minted by NewToken always carry exp.
		jwt.WithExpirationRequired(),
	}
	if j.issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(j.issuer))
	}
	if j.audience != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(j.audience))
	}

	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	}, parserOptions...)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return "", &Error{Err: ErrTokenExpired}
		default:
			return "", &Error{Err: ErrInvalidToken}
		}
	}

	if claims.Subject == "" {
		return "", &Error{Err: ErrInvalidClaims}
	}

	return claims.Subject, nil
}

// NewToken mints a signed token for the given subject.
func (j *JWT) NewToken(subject string) (string, error) {
	if subject == "" {
		return "", &Error{Err: ErrEmptySubject}
	}

	now := j.now()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
	}
	if j.issuer != "" {
		claims.Issuer = j.issuer
	}
	if j.audience != "" {
		claims.Audience = jwt.ClaimStrings{j.audience}
	}

	token := jwt.NewWithClaims(jwtSigningMethod, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return signed, nil
}

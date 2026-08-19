package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSecret = []byte("super-secret-signing-key")

// fixedClock returns a deterministic time source so token expiry is reproducible.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNewJWT_RequiresSecret(t *testing.T) {
	j, err := NewJWT(nil)
	assert.Nil(t, j)
	assert.Error(t, err)

	j, err = NewJWT([]byte{})
	assert.Nil(t, j)
	assert.Error(t, err)
}

func TestJWT_RoundTrip(t *testing.T) {
	j, err := NewJWT(testSecret)
	require.NoError(t, err)

	token, err := j.NewToken("user-123")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	subject, err := j.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", subject)
}

func TestJWT_NewTokenRequiresSubject(t *testing.T) {
	j, err := NewJWT(testSecret)
	require.NoError(t, err)

	_, err = j.NewToken("")
	assert.ErrorIs(t, err, ErrEmptySubject)
}

func TestJWT_ValidateEmptyToken(t *testing.T) {
	j, err := NewJWT(testSecret)
	require.NoError(t, err)

	_, err = j.ValidateToken("")
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestJWT_ValidateExpiredToken(t *testing.T) {
	issuedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	j, err := NewJWT(testSecret, WithTTL(time.Minute), withClock(fixedClock(issuedAt)))
	require.NoError(t, err)

	token, err := j.NewToken("user-123")
	require.NoError(t, err)

	// Validate from a clock well past the token's expiry.
	later := issuedAt.Add(time.Hour)
	expiredJWT, err := NewJWT(testSecret, withClock(fixedClock(later)))
	require.NoError(t, err)

	_, err = expiredJWT.ValidateToken(token)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestJWT_ValidateWrongSecret(t *testing.T) {
	signer, err := NewJWT(testSecret)
	require.NoError(t, err)
	token, err := signer.NewToken("user-123")
	require.NoError(t, err)

	verifier, err := NewJWT([]byte("a-different-secret"))
	require.NoError(t, err)

	_, err = verifier.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWT_RejectsNoneAlgorithm(t *testing.T) {
	// Forge a token signed with "none" — the classic algorithm-confusion attack.
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{Subject: "attacker"})
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	j, err := NewJWT(testSecret)
	require.NoError(t, err)

	_, err = j.ValidateToken(signed)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWT_IssuerAndAudience(t *testing.T) {
	j, err := NewJWT(testSecret, WithIssuer("go-starter"), WithAudience("api"))
	require.NoError(t, err)

	token, err := j.NewToken("user-123")
	require.NoError(t, err)

	subject, err := j.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", subject)

	// A validator expecting a different issuer must reject the token.
	other, err := NewJWT(testSecret, WithIssuer("someone-else"))
	require.NoError(t, err)
	_, err = other.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWT_MissingSubjectClaim(t *testing.T) {
	// Token with no subject, signed with the right key/alg.
	raw := jwt.NewWithClaims(jwtSigningMethod, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	signed, err := raw.SignedString(testSecret)
	require.NoError(t, err)

	j, err := NewJWT(testSecret)
	require.NoError(t, err)

	_, err = j.ValidateToken(signed)
	assert.ErrorIs(t, err, ErrInvalidClaims)
}

func TestJWT_RejectsTokenWithoutExpiry(t *testing.T) {
	// A token signed with the right key/alg and a subject but no "exp" claim
	// must be rejected — otherwise a forged token would never expire.
	raw := jwt.NewWithClaims(jwtSigningMethod, jwt.RegisteredClaims{
		Subject: "user-123",
	})
	signed, err := raw.SignedString(testSecret)
	require.NoError(t, err)

	j, err := NewJWT(testSecret)
	require.NoError(t, err)

	_, err = j.ValidateToken(signed)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWT_ImplementsInterfaces(t *testing.T) {
	j, err := NewJWT(testSecret)
	require.NoError(t, err)

	var _ Generator = j
	var _ TokenValidator = j

	// The auth Error wrapper unwraps to the sentinel, exercised across the suite.
	var authErr *Error
	_, err = j.ValidateToken("not-a-jwt")
	assert.True(t, errors.As(err, &authErr))
}

// Package credential verifies a plaintext password against a user's stored
// bcrypt hash. It is a pure business capability: it has no HTTP, generated
// transport, or direct database dependency, and it never returns or logs a
// password hash.
//
// It reads authentication data through a narrow, package-private port so it
// can be tested with a fake and wired to any repository that can produce a
// domain.Credential — in this starter, internal/modules/user.Repository.
package credential

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/rfizzle/go-starter/internal/domain"
)

// ErrInvalidCredentials is returned when the email is unknown, the password
// does not match the stored hash, or the stored hash is malformed. All three
// cases return this same error and perform the same bcrypt comparison work,
// so a caller cannot use the response or its timing to learn whether an
// account exists.
var ErrInvalidCredentials = errors.New("invalid credentials")

// unknownAccountHash is a valid bcrypt hash of a fixed, non-secret value with
// no corresponding real account. Authenticate compares against it whenever
// the requested email has no matching credential, so a missing account costs
// the same bcrypt work as a wrong password.
var unknownAccountHash = mustHash("go-starter-credential-unknown-account")

// Authenticated is the identity returned by a successful Authenticate call.
// It never carries a password hash.
type Authenticated struct {
	UserID string
	Email  string
}

// credentialReader is the narrow contract Service needs from user storage:
// enough to verify a password, never a full profile. It is satisfied
// structurally by internal/modules/user.Repository's CredentialByEmail; a
// future composition root injects the concrete without this package
// importing it.
type credentialReader interface {
	CredentialByEmail(ctx context.Context, email string) (domain.Credential, error)
}

// Service verifies plaintext passwords against stored bcrypt hashes.
type Service struct {
	reader credentialReader
}

// NewService builds a Service backed by reader. It returns an error if
// reader is nil so misconfiguration fails at construction rather than on the
// first authentication attempt.
func NewService(reader credentialReader) (*Service, error) {
	if reader == nil {
		return nil, errors.New("credential: reader is required")
	}
	return &Service{reader: reader}, nil
}

// Authenticate verifies password for email. It returns the authenticated
// identity on success, or ErrInvalidCredentials if the email is unknown, the
// password is wrong, or the stored hash is malformed. Any other error means
// the check could not be completed — a cancelled/expired context or a
// repository failure — and is wrapped for inspection with errors.Is/As; it
// is never converted to ErrInvalidCredentials. Context cancellation is
// honored both before the read and after the bcrypt comparison, so a reader
// that ignores cancellation and returns a stale successful result cannot
// produce a successful authentication.
func (s *Service) Authenticate(ctx context.Context, email, password string) (Authenticated, error) {
	if err := ctx.Err(); err != nil {
		return Authenticated{}, fmt.Errorf("credential: %w", err)
	}

	cred, err := s.reader.CredentialByEmail(ctx, email)
	switch {
	case err == nil:
		match := bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(password)) == nil
		if ctxErr := ctx.Err(); ctxErr != nil {
			return Authenticated{}, fmt.Errorf("credential: %w", ctxErr)
		}
		if !match {
			return Authenticated{}, ErrInvalidCredentials
		}
		return Authenticated{UserID: cred.UserID, Email: cred.Email}, nil
	case errors.Is(err, domain.ErrNotFound):
		_ = bcrypt.CompareHashAndPassword([]byte(unknownAccountHash), []byte(password))
		if ctxErr := ctx.Err(); ctxErr != nil {
			return Authenticated{}, fmt.Errorf("credential: %w", ctxErr)
		}
		return Authenticated{}, ErrInvalidCredentials
	default:
		return Authenticated{}, fmt.Errorf("credential: read credential: %w", err)
	}
}

// mustHash hashes password at package initialization. It panics on failure,
// which is unreachable for a fixed, valid-length input and a valid cost.
func mustHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("credential: failed to precompute unknown-account hash: " + err.Error())
	}
	return string(hash)
}

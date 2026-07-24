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
// does not match the stored hash, or the stored hash is malformed or uses an
// unexpected cost. All cases return this same error and perform one bcrypt
// comparison at the Service's configured cost.
var ErrInvalidCredentials = errors.New("invalid credentials")

const unknownAccountVerificationInput = "go-starter-credential-unknown-account"

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
	reader             credentialReader
	hashCost           int
	unknownAccountHash []byte
}

// NewService builds a Service backed by reader. hashCost is both the required
// cost of stored bcrypt hashes and the cost of the dummy comparison used for
// invalid credential paths. It returns an error for a nil reader or an invalid
// bcrypt cost so misconfiguration fails during construction.
func NewService(reader credentialReader, hashCost int) (*Service, error) {
	if reader == nil {
		return nil, errors.New("credential: reader is required")
	}
	if hashCost < bcrypt.MinCost || hashCost > bcrypt.MaxCost {
		return nil, fmt.Errorf(
			"credential: hash cost must be between %d and %d",
			bcrypt.MinCost,
			bcrypt.MaxCost,
		)
	}

	unknownAccountHash, err := bcrypt.GenerateFromPassword([]byte(unknownAccountVerificationInput), hashCost)
	if err != nil {
		return nil, fmt.Errorf("credential: create unknown-account hash: %w", err)
	}

	return &Service{
		reader:             reader,
		hashCost:           hashCost,
		unknownAccountHash: unknownAccountHash,
	}, nil
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
	comparisonHash := s.unknownAccountHash
	credentialUsable := false
	switch {
	case err == nil:
		storedHash := []byte(cred.PasswordHash)
		cost, costErr := bcrypt.Cost(storedHash)
		if costErr == nil && cost == s.hashCost {
			comparisonHash = storedHash
			credentialUsable = true
		}
	case errors.Is(err, domain.ErrNotFound):
		// Keep the dummy hash selected so missing accounts perform the same
		// comparison as malformed or unexpected-cost stored hashes.
	default:
		return Authenticated{}, fmt.Errorf("credential: read credential: %w", err)
	}

	match := bcrypt.CompareHashAndPassword(comparisonHash, []byte(password)) == nil
	if ctxErr := ctx.Err(); ctxErr != nil {
		return Authenticated{}, fmt.Errorf("credential: %w", ctxErr)
	}
	if !credentialUsable || !match {
		return Authenticated{}, ErrInvalidCredentials
	}

	return Authenticated{UserID: cred.UserID, Email: cred.Email}, nil
}

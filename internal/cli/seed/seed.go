// Package seed provides idempotent development database seeding for the
// starter. Callers must provide every record explicitly; this package contains
// no built-in identities or credentials.
//
// Seeding is idempotent: every run checks for an existing record before
// inserting, so running the `seed` subcommand repeatedly — or on a database
// that already has data — neither errors nor creates duplicates. This is the
// pattern real domains should copy: seed by natural key, skip what exists.
package seed

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/rfizzle/go-starter/internal/domain"
)

// UserStore is the slice of the user repository the seeder needs. Depending on
// this narrow interface rather than the concrete *user.Repository keeps the
// seeder trivially testable with a fake and documents exactly what seeding
// touches.
type UserStore interface {
	// ByEmail returns the user with the given email, or domain.ErrNotFound when
	// none exists.
	ByEmail(ctx context.Context, email string) (domain.User, error)
	// Create inserts a user and returns the stored row.
	Create(ctx context.Context, params domain.CreateUserParams) (domain.User, error)
}

// UserSeed is a single development user to seed. Password is plaintext and is
// bcrypt-hashed at seed time; it is never stored in the clear.
type UserSeed struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// Result reports observable seeding outcomes without exposing credentials.
type Result struct {
	Created  int
	Existing int
}

// Seeder inserts baseline records idempotently. Construct it with New.
type Seeder struct {
	users      UserStore
	logger     *zap.Logger
	bcryptCost int
	userSeeds  []UserSeed
}

// Option configures a Seeder.
type Option func(*Seeder)

// WithBcryptCost overrides the bcrypt cost used to hash seed passwords. Defaults
// to bcrypt.DefaultCost. Primarily a testing hook so tests can hash with the
// minimum cost.
func WithBcryptCost(cost int) Option {
	return func(s *Seeder) {
		s.bcryptCost = cost
	}
}

// New builds a Seeder over the given user store and explicit development
// records, logging progress (never credentials) through logger. It rejects a
// missing store or logger and missing records, identities, and credentials so
// misconfiguration fails before database work starts.
func New(users UserStore, logger *zap.Logger, userSeeds []UserSeed, options ...Option) (*Seeder, error) {
	if users == nil {
		return nil, fmt.Errorf("user store is nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}
	if len(userSeeds) == 0 {
		return nil, fmt.Errorf("at least one user seed is required")
	}
	for i, userSeed := range userSeeds {
		if userSeed.Email == "" {
			return nil, fmt.Errorf("user seed %d email is empty", i+1)
		}
		if userSeed.Password == "" {
			return nil, fmt.Errorf("user seed %d password is empty", i+1)
		}
	}

	s := &Seeder{
		users:      users,
		logger:     logger,
		bcryptCost: bcrypt.DefaultCost,
		userSeeds:  userSeeds,
	}

	for _, option := range options {
		option(s)
	}

	return s, nil
}

// Run seeds all configured records and reports how many were created or already
// existed. It is safe to call repeatedly: existing records are skipped, so only
// missing rows are inserted. The first error aborts the run and is returned.
func (s *Seeder) Run(ctx context.Context) (Result, error) {
	return s.seedUsers(ctx)
}

// seedUsers inserts each configured user that does not already exist, keyed on
// email. An existing user is left untouched (seeding never overwrites), so a
// changed seed password does not silently rewrite a real account.
func (s *Seeder) seedUsers(ctx context.Context) (Result, error) {
	var result Result
	for _, u := range s.userSeeds {
		_, err := s.users.ByEmail(ctx, u.Email)
		switch {
		case err == nil:
			// Already present; seeding is idempotent, so skip.
			s.logger.Debug("seed user already exists, skipping", zap.String("email", u.Email))
			result.Existing++
			continue
		case errors.Is(err, domain.ErrNotFound):
			// Fall through to create below.
		default:
			return result, fmt.Errorf("lookup seed user %q: %w", u.Email, err)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), s.bcryptCost)
		if err != nil {
			return result, fmt.Errorf("hash seed user %q password: %w", u.Email, err)
		}

		if _, err := s.users.Create(ctx, domain.CreateUserParams{
			Email:        u.Email,
			PasswordHash: string(hash),
			FirstName:    u.FirstName,
			LastName:     u.LastName,
		}); err != nil {
			return result, fmt.Errorf("create seed user %q: %w", u.Email, err)
		}

		s.logger.Info("seeded user", zap.String("email", u.Email))
		result.Created++
	}

	s.logger.Info("user seeding complete",
		zap.Int("created", result.Created),
		zap.Int("existing", result.Existing),
		zap.Int("total", len(s.userSeeds)),
	)
	return result, nil
}

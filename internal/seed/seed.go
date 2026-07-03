// Package seed provides idempotent database seeding for the starter. It inserts
// a baseline set of records (an example user) so a freshly-migrated database is
// immediately usable in development, demos, and integration tests.
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

	"github.com/rfizzle/go-starter/internal/storage"
	"github.com/rfizzle/go-starter/internal/user"
)

// UserStore is the slice of the user repository the seeder needs. Depending on
// this narrow interface rather than the concrete *user.Repository keeps the
// seeder trivially testable with a fake and documents exactly what seeding
// touches.
type UserStore interface {
	// ByEmail returns the user with the given email, or storage.ErrNotFound when
	// none exists.
	ByEmail(ctx context.Context, email string) (user.User, error)
	// Create inserts a user and returns the stored row.
	Create(ctx context.Context, params user.CreateParams) (user.User, error)
}

// UserSeed is a single baseline user to seed. Password is plaintext and is
// bcrypt-hashed at seed time; it is never stored in the clear.
type UserSeed struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// DefaultUsers is the baseline user set seeded when no explicit set is provided.
// It is intentionally a single, obviously-non-production account so the starter
// has a known login for local development without shipping real credentials.
var DefaultUsers = []UserSeed{
	{
		Email:     "admin@example.com",
		Password:  "password",
		FirstName: "Admin",
		LastName:  "User",
	},
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

// WithLogger sets the logger used for seeding progress. Defaults to a no-op
// logger.
func WithLogger(logger *zap.Logger) Option {
	return func(s *Seeder) {
		if logger != nil {
			s.logger = logger
		}
	}
}

// WithUsers overrides the set of users to seed. Defaults to DefaultUsers.
func WithUsers(users []UserSeed) Option {
	return func(s *Seeder) {
		s.userSeeds = users
	}
}

// WithBcryptCost overrides the bcrypt cost used to hash seed passwords. Defaults
// to bcrypt.DefaultCost. Primarily a testing hook so tests can hash with the
// minimum cost.
func WithBcryptCost(cost int) Option {
	return func(s *Seeder) {
		s.bcryptCost = cost
	}
}

// New builds a Seeder over the given user store. It returns an error when the
// store is nil so misconfiguration fails loudly at construction rather than
// mid-seed.
func New(users UserStore, options ...Option) (*Seeder, error) {
	if users == nil {
		return nil, fmt.Errorf("user store is nil")
	}

	s := &Seeder{
		users:      users,
		logger:     zap.NewNop(),
		bcryptCost: bcrypt.DefaultCost,
		userSeeds:  DefaultUsers,
	}

	for _, option := range options {
		option(s)
	}

	return s, nil
}

// Run seeds all configured records. It is safe to call repeatedly: existing
// records are skipped, so only missing rows are inserted. The first error
// aborts the run and is returned.
func (s *Seeder) Run(ctx context.Context) error {
	return s.seedUsers(ctx)
}

// seedUsers inserts each configured user that does not already exist, keyed on
// email. An existing user is left untouched (seeding never overwrites), so a
// changed seed password does not silently rewrite a real account.
func (s *Seeder) seedUsers(ctx context.Context) error {
	created := 0
	for _, u := range s.userSeeds {
		_, err := s.users.ByEmail(ctx, u.Email)
		switch {
		case err == nil:
			// Already present; seeding is idempotent, so skip.
			s.logger.Debug("seed user already exists, skipping", zap.String("email", u.Email))
			continue
		case errors.Is(err, storage.ErrNotFound):
			// Fall through to create below.
		default:
			return fmt.Errorf("lookup seed user %q: %w", u.Email, err)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), s.bcryptCost)
		if err != nil {
			return fmt.Errorf("hash seed user %q password: %w", u.Email, err)
		}

		if _, err := s.users.Create(ctx, user.CreateParams{
			Email:        u.Email,
			PasswordHash: string(hash),
			FirstName:    u.FirstName,
			LastName:     u.LastName,
		}); err != nil {
			return fmt.Errorf("create seed user %q: %w", u.Email, err)
		}

		s.logger.Info("seeded user", zap.String("email", u.Email))
		created++
	}

	s.logger.Info("user seeding complete",
		zap.Int("created", created),
		zap.Int("total", len(s.userSeeds)),
	)
	return nil
}

// Package domain defines business data, shared repository contracts, and
// business-facing errors. It is an infrastructure-free leaf package.
package domain

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound reports that a requested domain resource does not exist.
var ErrNotFound = errors.New("resource not found")

// User is the business representation of a user. Persistence-only fields such
// as password hashes must never be added to this type.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserParams contains the values required to persist a user.
// PasswordHash is an already-hashed credential; repositories do not hash it.
type CreateUserParams struct {
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
}

// UserFilter narrows a user list query. Zero values omit the corresponding
// filter. A zero Limit means no explicit result limit.
type UserFilter struct {
	EmailLike string
	Limit     uint64
}

// UserRepository is the shared persistence contract for users.
type UserRepository interface {
	ByID(ctx context.Context, id string) (User, error)
	ByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, params CreateUserParams) (User, error)
	List(ctx context.Context, filter UserFilter) ([]User, error)
}

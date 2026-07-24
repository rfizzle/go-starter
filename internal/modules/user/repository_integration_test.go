//go:build integration

package user_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rfizzle/go-starter/internal/domain"
	"github.com/rfizzle/go-starter/internal/modules/user"
	"github.com/rfizzle/go-starter/internal/platform/database/postgrestest"
)

func TestRepository_CreateAndReadAgainstPostgres(t *testing.T) {
	database := postgrestest.New(t).Database(t)
	repository := user.NewRepository(database.Pool)

	created, err := repository.Create(t.Context(), domain.CreateUserParams{
		Email:        "ada@example.com",
		PasswordHash: "already-hashed-password",
		FirstName:    "Ada",
		LastName:     "Lovelace",
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())

	byID, err := repository.ByID(t.Context(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, byID)

	byEmail, err := repository.ByEmail(t.Context(), created.Email)
	require.NoError(t, err)
	assert.Equal(t, created, byEmail)
}

func TestRepository_CredentialByEmailAgainstPostgres(t *testing.T) {
	database := postgrestest.New(t).Database(t)
	repository := user.NewRepository(database.Pool)

	created, err := repository.Create(t.Context(), domain.CreateUserParams{
		Email:        "credential@example.com",
		PasswordHash: "stored-password-hash",
		FirstName:    "Ada",
		LastName:     "Lovelace",
	})
	require.NoError(t, err)

	credential, err := repository.CredentialByEmail(t.Context(), created.Email)
	require.NoError(t, err)
	assert.Equal(t, created.ID, credential.UserID)
	assert.Equal(t, created.Email, credential.Email)
	assert.Equal(t, "stored-password-hash", credential.PasswordHash)
}

func TestRepository_CredentialByEmailNotFoundAgainstPostgres(t *testing.T) {
	database := postgrestest.New(t).Database(t)
	repository := user.NewRepository(database.Pool)

	_, err := repository.CredentialByEmail(t.Context(), "missing@example.com")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestRepository_ListUsesDynamicFiltersAgainstPostgres(t *testing.T) {
	database := postgrestest.New(t).Database(t)
	repository := user.NewRepository(database.Pool)
	base := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)

	err := database.ApplyFixture(t.Context(),
		postgrestest.Statement{
			SQL:  "INSERT INTO users (id, email, password_hash, first_name, last_name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			Args: []any{"00000000-0000-0000-0000-000000000001", "old@example.com", "hash", "Old", "User", base, base},
		},
		postgrestest.Statement{
			SQL:  "INSERT INTO users (id, email, password_hash, first_name, last_name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			Args: []any{"00000000-0000-0000-0000-000000000002", "new@example.com", "hash", "New", "User", base.Add(time.Minute), base.Add(time.Minute)},
		},
		postgrestest.Statement{
			SQL:  "INSERT INTO users (id, email, password_hash, first_name, last_name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			Args: []any{"00000000-0000-0000-0000-000000000003", "other@invalid.test", "hash", "Other", "User", base.Add(2 * time.Minute), base.Add(2 * time.Minute)},
		},
	)
	require.NoError(t, err)

	users, err := repository.List(t.Context(), domain.UserFilter{
		EmailLike: "EXAMPLE.COM",
		Limit:     1,
	})
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000002", users[0].ID)
	assert.Equal(t, "new@example.com", users[0].Email)
}

func TestRepository_CreatePreservesUniqueConstraintError(t *testing.T) {
	database := postgrestest.New(t).Database(t)
	repository := user.NewRepository(database.Pool)
	params := domain.CreateUserParams{Email: "duplicate@example.com", PasswordHash: "hash"}

	_, err := repository.Create(t.Context(), params)
	require.NoError(t, err)

	_, err = repository.Create(t.Context(), params)
	require.Error(t, err)

	var postgresError *pgconn.PgError
	require.True(t, errors.As(err, &postgresError))
	assert.Equal(t, "23505", postgresError.Code)
	assert.Equal(t, "users_email_key", postgresError.ConstraintName)
}

func TestRepository_WithinTxCommitsAndRollsBackAgainstPostgres(t *testing.T) {
	database := postgrestest.New(t).Database(t)
	repository := user.NewRepository(database.Pool)

	err := repository.WithinTx(t.Context(), func(txRepository *user.Repository) error {
		_, err := txRepository.Create(t.Context(), domain.CreateUserParams{
			Email:        "committed@example.com",
			PasswordHash: "hash",
		})
		return err
	})
	require.NoError(t, err)

	_, err = repository.ByEmail(t.Context(), "committed@example.com")
	require.NoError(t, err)

	rollbackCause := errors.New("force transaction rollback")
	err = repository.WithinTx(t.Context(), func(txRepository *user.Repository) error {
		_, createErr := txRepository.Create(t.Context(), domain.CreateUserParams{
			Email:        "rolled-back@example.com",
			PasswordHash: "hash",
		})
		if createErr != nil {
			return createErr
		}
		return rollbackCause
	})
	require.ErrorIs(t, err, rollbackCause)

	_, err = repository.ByEmail(t.Context(), "rolled-back@example.com")
	require.ErrorIs(t, err, domain.ErrNotFound)
}

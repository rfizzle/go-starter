//go:build integration

package seed_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/rfizzle/go-starter/internal/cli/seed"
	"github.com/rfizzle/go-starter/internal/modules/user"
	"github.com/rfizzle/go-starter/internal/platform/database/postgrestest"
)

func TestMain(m *testing.M) {
	os.Exit(postgrestest.RunMain(m))
}

func TestSeeder_RunIsIdempotentAgainstPostgres(t *testing.T) {
	database := postgrestest.Shared(t).Database(t)
	repository := user.NewRepository(database.Pool)
	firstUsers := []seed.UserSeed{{
		Email:     "seeded@example.com",
		Password:  "first-development-password",
		FirstName: "Seeded",
		LastName:  "User",
	}}
	seeder, err := seed.New(
		repository,
		zap.NewNop(),
		firstUsers,
		seed.WithBcryptCost(bcrypt.MinCost),
	)
	require.NoError(t, err)

	firstResult, err := seeder.Run(t.Context())
	require.NoError(t, err)
	assert.Equal(t, seed.Result{Created: 1}, firstResult)

	storedBefore, err := repository.CredentialByEmail(t.Context(), firstUsers[0].Email)
	require.NoError(t, err)
	require.NoError(
		t,
		bcrypt.CompareHashAndPassword([]byte(storedBefore.PasswordHash), []byte(firstUsers[0].Password)),
	)

	secondUsers := []seed.UserSeed{{
		Email:     firstUsers[0].Email,
		Password:  "replacement-password",
		FirstName: "Changed",
		LastName:  "Name",
	}}
	secondSeeder, err := seed.New(
		repository,
		zap.NewNop(),
		secondUsers,
		seed.WithBcryptCost(bcrypt.MinCost),
	)
	require.NoError(t, err)
	secondResult, err := secondSeeder.Run(t.Context())
	require.NoError(t, err)
	assert.Equal(t, seed.Result{Existing: 1}, secondResult)

	seeded, err := repository.ByEmail(t.Context(), firstUsers[0].Email)
	require.NoError(t, err)
	assert.Equal(t, firstUsers[0].FirstName, seeded.FirstName)
	assert.Equal(t, firstUsers[0].LastName, seeded.LastName)

	storedAfter, err := repository.CredentialByEmail(t.Context(), firstUsers[0].Email)
	require.NoError(t, err)
	assert.Equal(t, storedBefore.PasswordHash, storedAfter.PasswordHash)
	assert.Error(
		t,
		bcrypt.CompareHashAndPassword([]byte(storedAfter.PasswordHash), []byte(secondUsers[0].Password)),
	)

	var count int
	err = database.Pool.QueryRow(
		t.Context(),
		"SELECT count(*) FROM users WHERE email = $1",
		firstUsers[0].Email,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

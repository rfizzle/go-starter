//go:build integration

package seed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/rfizzle/go-starter/internal/cli/seed"
	"github.com/rfizzle/go-starter/internal/modules/user"
	"github.com/rfizzle/go-starter/internal/platform/database/postgrestest"
)

func TestSeeder_RunIsIdempotentAgainstPostgres(t *testing.T) {
	database := postgrestest.New(t).Database(t)
	repository := user.NewRepository(database.Pool)
	users := []seed.UserSeed{{
		Email:     "seeded@example.com",
		Password:  "development-only-password",
		FirstName: "Seeded",
		LastName:  "User",
	}}
	seeder, err := seed.New(
		repository,
		seed.WithUsers(users),
		seed.WithBcryptCost(bcrypt.MinCost),
	)
	require.NoError(t, err)

	require.NoError(t, seeder.Run(t.Context()))
	require.NoError(t, seeder.Run(t.Context()))

	seeded, err := repository.ByEmail(t.Context(), users[0].Email)
	require.NoError(t, err)
	assert.Equal(t, users[0].FirstName, seeded.FirstName)
	assert.Equal(t, users[0].LastName, seeded.LastName)

	var count int
	err = database.Pool.QueryRow(
		t.Context(),
		"SELECT count(*) FROM users WHERE email = $1",
		users[0].Email,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

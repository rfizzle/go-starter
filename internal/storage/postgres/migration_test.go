package postgres

import (
	"context"
	"io/fs"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/rfizzle/go-starter/internal/storage"
)

// TestMigrateRequiresPool verifies migrations fail fast on a client that has
// no underlying pool (e.g. connection- or mock-backed), since goose runs
// through a database/sql handle wrapped over the pgx pool.
func TestMigrateRequiresPool(t *testing.T) {
	client := &Client{logger: zaptest.NewLogger(t)}

	err := client.Migrate(Migrations)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pool-backed client")
}

// TestMigrationStatusRequiresPool mirrors TestMigrateRequiresPool for the
// status path and confirms it reports StatusNone alongside the error.
func TestMigrationStatusRequiresPool(t *testing.T) {
	client := &Client{logger: zaptest.NewLogger(t)}

	status, err := client.MigrationStatus(context.Background(), Migrations)
	require.Error(t, err)
	assert.Equal(t, storage.StatusNone, status)
}

// TestEmbeddedMigrationsAreValidGoose ensures every embedded migration lives in
// the expected sub-directory, carries goose Up/Down annotations, and has a
// numeric version prefix. This guards the schema source that both the runtime
// migrator and sqlc read from.
func TestEmbeddedMigrationsAreValidGoose(t *testing.T) {
	sub, err := fs.Sub(Migrations, migrationsDir)
	require.NoError(t, err)

	files, err := fs.Glob(sub, "*.sql")
	require.NoError(t, err)
	require.NotEmpty(t, files, "expected at least one embedded migration")

	for _, f := range files {
		data, err := fs.ReadFile(sub, f)
		require.NoError(t, err)
		content := string(data)
		assert.Contains(t, content, "-- +goose Up", "%s missing goose Up annotation", f)
		assert.Contains(t, content, "-- +goose Down", "%s missing goose Down annotation", f)

		// Each filename must start with a numeric version so goose orders it.
		_, err = goose.NumericComponent(f)
		assert.NoError(t, err, "%s must have a numeric version prefix", f)
	}

	// Filenames must be strictly increasing so version ordering is unambiguous.
	assert.True(t, isStrictlySorted(files), "migration files must sort in version order: %v", files)
}

func isStrictlySorted(files []string) bool {
	for i := 1; i < len(files); i++ {
		if strings.Compare(files[i-1], files[i]) >= 0 {
			return false
		}
	}
	return true
}

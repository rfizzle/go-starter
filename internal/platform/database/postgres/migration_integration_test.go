//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap/zaptest"

	"github.com/rfizzle/go-starter/internal/platform/database/postgres"
	"github.com/rfizzle/go-starter/internal/platform/database/postgrestest"
)

func TestMain(m *testing.M) {
	os.Exit(postgrestest.RunMain(m))
}

func TestMigrateEmptyDatabaseToCurrentSchemaIsIdempotent(t *testing.T) {
	database := postgrestest.Shared(t).EmptyDatabase(t)
	logger := zaptest.NewLogger(t)

	for run := 1; run <= 2; run++ {
		if err := postgres.Migrate(t.Context(), database.Pool, logger, postgres.Migrations); err != nil {
			t.Fatalf("Migrate() run %d error: %v", run, err)
		}
	}

	var (
		usersTableExists  bool
		appliedMigrations int
	)
	if err := database.Pool.QueryRow(
		t.Context(),
		"SELECT to_regclass('public.users') IS NOT NULL",
	).Scan(&usersTableExists); err != nil {
		t.Fatalf("query users table: %v", err)
	}
	if err := database.Pool.QueryRow(
		t.Context(),
		"SELECT count(*) FROM goose_db_version WHERE is_applied AND version_id > 0",
	).Scan(&appliedMigrations); err != nil {
		t.Fatalf("count applied migrations: %v", err)
	}

	if !usersTableExists {
		t.Fatal("users table does not exist after migration")
	}
	if appliedMigrations != 2 {
		t.Fatalf("applied migrations = %d, want 2", appliedMigrations)
	}
}

func TestMigrateSerializesConcurrentPostgresAttempts(t *testing.T) {
	database := postgrestest.Shared(t).EmptyDatabase(t)
	testMigrations := fstest.MapFS{
		"migrations/00001_serialization_probe.sql": {
			Data: []byte(`-- +goose Up
-- +goose StatementBegin
SELECT pg_sleep(1);
CREATE TABLE migration_serialization_probe (id integer PRIMARY KEY);
-- +goose StatementEnd

-- +goose Down
DROP TABLE migration_serialization_probe;
`),
		},
	}

	pools := make([]*pgxpool.Pool, 2)
	for i := range pools {
		pool, err := pgxpool.New(t.Context(), database.URI)
		if err != nil {
			t.Fatalf("pgxpool.New() pool %d error: %v", i, err)
		}
		t.Cleanup(pool.Close)
		pools[i] = pool
	}

	logger := zaptest.NewLogger(t)
	start := make(chan struct{})
	results := make(chan error, len(pools))
	var workers sync.WaitGroup
	for _, pool := range pools {
		workers.Go(func() {
			<-start
			results <- postgres.MigrateFS(t.Context(), pool, logger, testMigrations)
		})
	}
	close(start)
	workers.Wait()
	close(results)

	for err := range results {
		if err != nil {
			t.Errorf("concurrent migrate() error: %v", err)
		}
	}

	var exists bool
	if err := database.Pool.QueryRow(
		t.Context(),
		"SELECT to_regclass('public.migration_serialization_probe') IS NOT NULL",
	).Scan(&exists); err != nil {
		t.Fatalf("query migrated table: %v", err)
	}
	if !exists {
		t.Fatal("migration_serialization_probe table does not exist")
	}
}

func TestMigrateHonorsCanceledContext(t *testing.T) {
	database := postgrestest.Shared(t).EmptyDatabase(t)
	pool, err := pgxpool.New(t.Context(), database.URI)
	if err != nil {
		t.Fatalf("pgxpool.New() error: %v", err)
	}
	t.Cleanup(pool.Close)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err = postgres.Migrate(ctx, pool, zaptest.NewLogger(t), postgres.Migrations)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Migrate() error = %v, want context cancellation", err)
	}
}

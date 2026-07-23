//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap/zaptest"
)

var integrationDatabaseSequence atomic.Uint64

func TestMigrateSerializesConcurrentPostgresAttempts(t *testing.T) {
	testURI := createMigrationTestDatabase(t)

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

	clients := make([]*Client, 2)
	for i := range clients {
		pool, err := pgxpool.New(t.Context(), testURI)
		if err != nil {
			t.Fatalf("pgxpool.New() client %d error: %v", i, err)
		}
		t.Cleanup(pool.Close)

		client, err := Open(
			t.Context(),
			WithPool(pool),
			WithLogger(zaptest.NewLogger(t)),
		)
		if err != nil {
			t.Fatalf("Open() client %d error: %v", i, err)
		}
		clients[i] = client
	}

	start := make(chan struct{})
	results := make(chan error, len(clients))
	var workers sync.WaitGroup
	for _, client := range clients {
		workers.Go(func() {
			<-start
			results <- client.migrate(t.Context(), testMigrations)
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
	if err := clients[0].pool.QueryRow(
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
	testURI := createMigrationTestDatabase(t)
	pool, err := pgxpool.New(t.Context(), testURI)
	if err != nil {
		t.Fatalf("pgxpool.New() error: %v", err)
	}
	t.Cleanup(pool.Close)

	client, err := Open(
		t.Context(),
		WithPool(pool),
		WithLogger(zaptest.NewLogger(t)),
	)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err = client.Migrate(ctx, Migrations)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Migrate() error = %v, want context cancellation", err)
	}
}

func createMigrationTestDatabase(t *testing.T) string {
	t.Helper()

	adminURI := os.Getenv("GOSTARTER_TEST_DATABASE_URI")
	if adminURI == "" {
		t.Skip("GOSTARTER_TEST_DATABASE_URI is not set")
	}

	adminConfig, err := pgx.ParseConfig(adminURI)
	if err != nil {
		t.Fatalf("pgx.ParseConfig() error: %v", err)
	}
	admin, err := pgx.ConnectConfig(t.Context(), adminConfig)
	if err != nil {
		t.Fatalf("connect test database administrator: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := admin.Close(cleanupCtx); err != nil {
			t.Errorf("close test database administrator: %v", err)
		}
	})

	databaseName := fmt.Sprintf(
		"gostarter_migration_%d_%d",
		os.Getpid(),
		integrationDatabaseSequence.Add(1),
	)
	identifier := pgx.Identifier{databaseName}.Sanitize()
	if _, err := admin.Exec(t.Context(), "CREATE DATABASE "+identifier); err != nil {
		t.Fatalf("create isolated migration database: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(
			cleanupCtx,
			"DROP DATABASE "+identifier+" WITH (FORCE)",
		); err != nil {
			t.Errorf("drop isolated migration database: %v", err)
		}
	})

	testConfig := adminConfig.Copy()
	testConfig.Database = databaseName
	return testConfig.ConnString()
}

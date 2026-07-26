//go:build integration

package postgrestest

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMain(m *testing.M) {
	os.Exit(RunMain(m))
}

// TestSharedReusesOneContainer protects the one-container-per-test-binary
// contract: a second call must hand back the running harness instead of
// starting another PostgreSQL container.
func TestSharedReusesOneContainer(t *testing.T) {
	first := Shared(t)
	if second := Shared(t); second != first {
		t.Fatalf("Shared() returned %p then %p, want one shared harness", first, second)
	}
}

func TestHarnessCreatesMigratedIsolatedDatabases(t *testing.T) {
	harness := Shared(t)
	first := harness.Database(t)
	second := harness.Database(t)

	var firstName, secondName string
	if err := first.Pool.QueryRow(t.Context(), "SELECT current_database()").Scan(&firstName); err != nil {
		t.Fatalf("query first database name: %v", err)
	}
	if err := second.Pool.QueryRow(t.Context(), "SELECT current_database()").Scan(&secondName); err != nil {
		t.Fatalf("query second database name: %v", err)
	}
	if firstName == secondName {
		t.Fatalf("isolated database names are both %q", firstName)
	}

	fixture := []Statement{
		{
			SQL: `INSERT INTO users (
				id, email, password_hash, first_name, last_name, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $6)`,
			Args: []any{
				"00000000-0000-0000-0000-000000000020",
				"fixture@example.com",
				"fixed-hash",
				"Fixture",
				"User",
				time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	if err := first.ApplyFixture(t.Context(), fixture...); err != nil {
		t.Fatalf("ApplyFixture() error: %v", err)
	}

	var email string
	if err := first.Pool.QueryRow(
		t.Context(),
		"SELECT email FROM users WHERE id = $1",
		"00000000-0000-0000-0000-000000000020",
	).Scan(&email); err != nil {
		t.Fatalf("query deterministic fixture: %v", err)
	}
	if email != "fixture@example.com" {
		t.Fatalf("fixture email = %q, want %q", email, "fixture@example.com")
	}

	var secondCount int
	if err := second.Pool.QueryRow(t.Context(), "SELECT count(*) FROM users").Scan(&secondCount); err != nil {
		t.Fatalf("query second isolated database: %v", err)
	}
	if secondCount != 0 {
		t.Fatalf("second isolated database user count = %d, want 0", secondCount)
	}
}

func TestHarnessSupportsEmptyDatabaseAndAtomicFixtures(t *testing.T) {
	harness := Shared(t)
	empty := harness.EmptyDatabase(t)

	var tableName string
	err := empty.Pool.QueryRow(t.Context(), "SELECT 'users'::regclass::text").Scan(&tableName)
	if err == nil || !isUndefinedTable(err) {
		t.Fatalf("empty database users lookup error = %v, want undefined table", err)
	}

	database := harness.Database(t)
	err = database.ApplyFixture(
		t.Context(),
		Statement{
			SQL:  "INSERT INTO users (email, password_hash) VALUES ($1, $2)",
			Args: []any{"rolled-back@example.com", "fixed-hash"},
		},
		Statement{SQL: "INSERT INTO missing_fixture_table (id) VALUES (1)"},
	)
	if err == nil {
		t.Fatal("ApplyFixture() error = nil, want failed statement")
	}

	var count int
	if err := database.Pool.QueryRow(
		t.Context(),
		"SELECT count(*) FROM users WHERE email = $1",
		"rolled-back@example.com",
	).Scan(&count); err != nil {
		t.Fatalf("query rolled-back fixture: %v", err)
	}
	if count != 0 {
		t.Fatalf("rolled-back fixture user count = %d, want 0", count)
	}
}

func isUndefinedTable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}

//go:build integration

package postgrestest

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/rfizzle/go-starter/internal/platform/database/postgres"
)

const (
	postgresImage = "postgres:17.5-alpine"
	databaseUser  = "gostarter_test"
	databasePass  = "gostarter_test"
	cleanupBudget = 10 * time.Second
)

// Harness owns one disposable PostgreSQL container and its administrator pool.
// Create isolated databases through Database or EmptyDatabase rather than
// connecting tests to the administrator database.
type Harness struct {
	adminURL  *url.URL
	adminPool *pgxpool.Pool
	sequence  atomic.Uint64
}

// Database is an isolated PostgreSQL database and its caller-owned pool. The
// harness registers pool and database cleanup with the testing.T passed to the
// creation method.
type Database struct {
	URI  string
	Pool *pgxpool.Pool
}

// Statement is one ordered fixture statement. Arguments are passed separately
// to pgx so fixture values remain parameterized.
type Statement struct {
	SQL  string
	Args []any
}

// New starts a disposable PostgreSQL container. When Docker is unavailable,
// local integration tests skip with Testcontainers' provider diagnostic.
// Setting GOSTARTER_REQUIRE_DOCKER=1 disables that skip so CI fails instead.
func New(t *testing.T) *Harness {
	t.Helper()

	if !dockerRequired() {
		testcontainers.SkipIfProviderIsNotHealthy(t)
	}

	startupCtx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	container, err := tcpostgres.Run(
		startupCtx,
		postgresImage,
		tcpostgres.WithDatabase("postgres"),
		tcpostgres.WithUsername(databaseUser),
		tcpostgres.WithPassword(databasePass),
		tcpostgres.BasicWaitStrategies(),
		testcontainers.WithLabels(map[string]string{
			"com.gostarter.test-harness": "postgres",
		}),
	)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start PostgreSQL test container: %v", err)
	}

	adminURI, err := container.ConnectionString(startupCtx, "sslmode=disable")
	if err != nil {
		t.Fatalf("resolve PostgreSQL test container connection string: %v", err)
	}
	adminURL, err := url.Parse(adminURI)
	if err != nil {
		t.Fatalf("parse PostgreSQL test container connection string: %v", err)
	}
	adminPool, err := pgxpool.New(startupCtx, adminURI)
	if err != nil {
		t.Fatalf("open PostgreSQL test administrator pool: %v", err)
	}
	t.Cleanup(adminPool.Close)
	if err := adminPool.Ping(startupCtx); err != nil {
		t.Fatalf("ping PostgreSQL test container: %v", err)
	}

	return &Harness{
		adminURL:  adminURL,
		adminPool: adminPool,
	}
}

// Database creates a migrated database isolated from every other harness
// database. Its pool and schema are removed automatically when t completes.
func (h *Harness) Database(t *testing.T) *Database {
	t.Helper()
	return h.database(t, true)
}

// EmptyDatabase creates an isolated database without applying migrations. Use
// it for tests whose subject is the migration lifecycle itself.
func (h *Harness) EmptyDatabase(t *testing.T) *Database {
	t.Helper()
	return h.database(t, false)
}

func (h *Harness) database(t *testing.T, migrate bool) *Database {
	t.Helper()

	databaseName := testDatabaseName(t.Name(), h.sequence.Add(1))
	identifier := pgx.Identifier{databaseName}.Sanitize()
	if _, err := h.adminPool.Exec(t.Context(), "CREATE DATABASE "+identifier); err != nil {
		t.Fatalf("create isolated PostgreSQL database %q: %v", databaseName, err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupBudget)
		defer cancel()
		if _, err := h.adminPool.Exec(
			cleanupCtx,
			"DROP DATABASE "+identifier+" WITH (FORCE)",
		); err != nil {
			t.Errorf("drop isolated PostgreSQL database %q: %v", databaseName, err)
		}
	})

	databaseURL := *h.adminURL
	databaseURL.Path = "/" + databaseName
	databaseURL.RawPath = ""
	databaseURI := databaseURL.String()
	pool, err := pgxpool.New(t.Context(), databaseURI)
	if err != nil {
		t.Fatalf("open isolated PostgreSQL database %q: %v", databaseName, err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(t.Context()); err != nil {
		t.Fatalf("ping isolated PostgreSQL database %q: %v", databaseName, err)
	}

	if migrate {
		client, err := postgres.Open(t.Context(), postgres.WithPool(pool))
		if err != nil {
			t.Fatalf("construct migration client for %q: %v", databaseName, err)
		}
		if err := client.Migrate(t.Context(), postgres.Migrations); err != nil {
			t.Fatalf("migrate isolated PostgreSQL database %q: %v", databaseName, err)
		}
	}

	return &Database{
		URI:  databaseURI,
		Pool: pool,
	}
}

// ApplyFixture executes statements in order inside one transaction. A failed
// statement rolls back every earlier statement in the fixture.
func (d *Database) ApplyFixture(ctx context.Context, statements ...Statement) error {
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin fixture transaction: %w", err)
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cleanupBudget)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()

	for i, statement := range statements {
		if strings.TrimSpace(statement.SQL) == "" {
			return fmt.Errorf("fixture statement %d is empty", i+1)
		}
		if _, err := tx.Exec(ctx, statement.SQL, statement.Args...); err != nil {
			return fmt.Errorf("execute fixture statement %d: %w", i+1, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit fixture transaction: %w", err)
	}
	return nil
}

func dockerRequired() bool {
	return strings.TrimSpace(strings.ToLower(os.Getenv("GOSTARTER_REQUIRE_DOCKER"))) == "1"
}

func testDatabaseName(testName string, sequence uint64) string {
	var name strings.Builder
	for _, r := range strings.ToLower(testName) {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			name.WriteRune(r)
		default:
			name.WriteByte('_')
		}
	}

	base := strings.Trim(name.String(), "_")
	if base == "" {
		base = "test"
	}
	if len(base) > 48 {
		base = base[:48]
	}
	return fmt.Sprintf("gs_%s_%d", base, sequence)
}

//go:build integration

package postgrestest

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
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
	startupBudget = 2 * time.Minute
)

// errProviderUnavailable reports that no healthy container provider exists.
// Tests skip on it unless GOSTARTER_REQUIRE_CONTAINERS=1 demands a container.
var errProviderUnavailable = errors.New("container provider is unavailable")

// Harness owns one disposable PostgreSQL container and its administrator pool.
// One harness serves a whole test binary; isolation comes from Database and
// EmptyDatabase, which create a distinct database per call. Never connect a
// test to the administrator database.
type Harness struct {
	adminURL  *url.URL
	adminPool *pgxpool.Pool
	sequence  atomic.Uint64
}

// shared is the test binary's single harness. It is started by the first
// Shared call and released by RunMain once the package's tests complete.
var shared struct {
	mu      sync.Mutex
	started bool
	harness *Harness
	stop    func()
	err     error
}

// RunMain runs a package's tests with one shared PostgreSQL container rather
// than one per test. Integration-test packages install it as their TestMain:
//
//	func TestMain(m *testing.M) {
//		os.Exit(postgrestest.RunMain(m))
//	}
//
// The container starts on the first Shared call, so a run that selects no
// harness-backed test starts nothing.
func RunMain(m *testing.M) int {
	defer stopShared()
	return m.Run()
}

// Shared returns the test binary's harness, starting its container on first
// use. The calling package must install RunMain as its TestMain. When no
// container engine is reachable the test skips, unless
// GOSTARTER_REQUIRE_CONTAINERS=1 makes an unusable provider fatal instead.
func Shared(t *testing.T) *Harness {
	t.Helper()

	shared.mu.Lock()
	defer shared.mu.Unlock()

	if !shared.started {
		shared.started = true
		shared.harness, shared.stop, shared.err = startShared()
	}
	if shared.err != nil {
		if errors.Is(shared.err, errProviderUnavailable) {
			t.Skipf("skipping integration test: %v", shared.err)
		}
		t.Fatalf("start shared PostgreSQL test container: %v", shared.err)
	}
	return shared.harness
}

func startShared() (*Harness, func(), error) {
	if !containersRequired() {
		if err := providerHealth(); err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errProviderUnavailable, err)
		}
	}

	startupCtx, cancel := context.WithTimeout(context.Background(), startupBudget)
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
	// TerminateContainer tolerates a nil container, so it also covers a
	// partially started container reported alongside an error.
	terminate := func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			fmt.Fprintf(os.Stderr, "terminate PostgreSQL test container: %v\n", err)
		}
	}
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("start PostgreSQL test container: %w", err)
	}

	adminURI, err := container.ConnectionString(startupCtx, "sslmode=disable")
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("resolve PostgreSQL test container connection string: %w", err)
	}
	adminURL, err := url.Parse(adminURI)
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("parse PostgreSQL test container connection string: %w", err)
	}
	adminPool, err := pgxpool.New(startupCtx, adminURI)
	if err != nil {
		terminate()
		return nil, nil, fmt.Errorf("open PostgreSQL test administrator pool: %w", err)
	}
	if err := adminPool.Ping(startupCtx); err != nil {
		adminPool.Close()
		terminate()
		return nil, nil, fmt.Errorf("ping PostgreSQL test container: %w", err)
	}

	harness := &Harness{
		adminURL:  adminURL,
		adminPool: adminPool,
	}
	return harness, func() {
		adminPool.Close()
		terminate()
	}, nil
}

func stopShared() {
	shared.mu.Lock()
	defer shared.mu.Unlock()

	if shared.stop != nil {
		shared.stop()
		shared.stop = nil
	}
	shared.harness = nil
}

// providerHealth reports whether a container provider can serve this binary.
// testcontainers speaks the Docker-compatible API that rootless Podman serves,
// so ProviderDocker covers the repository's Podman engine through DOCKER_HOST
// or the default socket. Provider construction panics on some misconfigured
// hosts, which must read as an unavailable provider rather than crash the
// test binary.
func providerHealth() (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("container provider panicked: %v", recovered)
		}
	}()

	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		return fmt.Errorf("resolve container provider: %w", err)
	}
	return provider.Health(context.Background())
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

func containersRequired() bool {
	return strings.TrimSpace(strings.ToLower(os.Getenv("GOSTARTER_REQUIRE_CONTAINERS"))) == "1"
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

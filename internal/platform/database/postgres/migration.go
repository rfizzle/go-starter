package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/internal/platform/database"
)

// migrationsDir is the sub-directory inside the embedded filesystem that holds
// the goose SQL migrations. The embed.FS carries the "migrations/" prefix (see
// embed.go), so goose is pointed at that sub-tree rather than the FS root.
const migrationsDir = "migrations"

// gooseDialect is the goose dialect used for the migration store. It selects
// the PostgreSQL versioning table implementation.
const gooseDialect = goose.DialectPostgres

// newGooseProvider builds a goose provider over the connection pool. goose
// speaks database/sql, so the pgx pool is wrapped with stdlib.OpenDBFromPool;
// closing that *sql.DB does not close the underlying pgxpool.Pool (it releases
// pooled connections back), so the caller-owned pool is left intact.
//
// It returns the provider, a cleanup func that must be called to release the
// wrapping *sql.DB, and any construction error.
func newGooseProvider(pool *pgxpool.Pool, migrationFS fs.FS) (*goose.Provider, func(), error) {
	if pool == nil {
		return nil, nil, fmt.Errorf("migrations require a non-nil pool")
	}

	// goose globs "*.sql" at the FS root, so hand it the migrations sub-tree
	// rather than the whole embed.FS (which nests files under "migrations/").
	sub, err := fs.Sub(migrationFS, migrationsDir)
	if err != nil {
		return nil, nil, fmt.Errorf("fs.Sub(%q): %w", migrationsDir, err)
	}

	db := stdlib.OpenDBFromPool(pool)

	// Hold one PostgreSQL session-level advisory lock for the entire Up
	// operation. Competing commands or startup migrators therefore serialize on
	// the database rather than racing to create the version table or schema.
	// Goose's locker retries for a bounded five minutes by default and stops
	// immediately when the operation context is cancelled.
	sessionLocker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("lock.NewPostgresSessionLocker(): %w", err)
	}

	// Goose only emits progress output when verbose, which is left off: the
	// applied-migration loop logs results through zap instead.
	gp, err := goose.NewProvider(
		gooseDialect,
		db,
		sub,
		goose.WithSessionLocker(sessionLocker),
	)
	if err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("goose.NewProvider(): %w", err)
	}

	return gp, func() { _ = db.Close() }, nil
}

// Migrate applies all pending, forward-only migrations from the embedded
// filesystem. goose tracks applied migrations by numeric version in its own
// table and runs each inside its own transaction. It is safe to call on every
// invocation: already-applied versions are skipped and a fully-migrated
// database is a no-op. A PostgreSQL session advisory lock serializes concurrent
// callers. Applied migrations are logged through logger; a nil logger
// discards them.
//
// goose keys purely on the version number and does not checksum migration
// bodies, so editing an already-applied file is silently ignored — change the
// schema by adding a new versioned migration, never by editing an applied one.
func Migrate(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger, migrationFS embed.FS) error {
	return migrate(ctx, pool, logger, migrationFS)
}

// migrate accepts fs.FS so integration tests can exercise PostgreSQL locking
// with a deliberately slow disposable migration while the public contract
// remains limited to the repository's embedded migration set.
func migrate(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger, migrationFS fs.FS) error {
	if logger == nil {
		logger = zap.NewNop()
	}

	gp, cleanup, err := newGooseProvider(pool, migrationFS)
	if err != nil {
		logger.Error("issue constructing migration provider", zap.Error(err))
		return err
	}
	defer cleanup()

	results, err := gp.Up(ctx)
	if err != nil {
		logger.Error("issue running migrations", zap.Error(err))
		return fmt.Errorf("goose up: %w", err)
	}

	for _, r := range results {
		logger.Info("applied migration", zap.String("migration", r.String()))
	}

	return nil
}

// MigrationStatus reports whether all, some, or none of the embedded
// migrations have been applied. It derives the coarse database.MigrationStatus
// from goose's per-migration status list.
//
// It is not strictly read-only: querying status ensures goose's version table
// exists, so calling it against a pristine database creates that table (with no
// migration rows). Keep this in mind before wiring it into a readiness probe.
func MigrationStatus(ctx context.Context, pool *pgxpool.Pool, migrationFS embed.FS) (database.MigrationStatus, error) {
	gp, cleanup, err := newGooseProvider(pool, migrationFS)
	if err != nil {
		return database.StatusNone, err
	}
	defer cleanup()

	statuses, err := gp.Status(ctx)
	if err != nil {
		return database.StatusNone, fmt.Errorf("goose status: %w", err)
	}

	total := len(statuses)
	applied := 0
	for _, s := range statuses {
		if s.State == goose.StateApplied {
			applied++
		}
	}

	switch applied {
	case 0:
		return database.StatusNone, nil
	case total:
		return database.StatusComplete, nil
	default:
		return database.StatusPartial, nil
	}
}

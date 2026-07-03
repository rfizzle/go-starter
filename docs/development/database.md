# Database Patterns Guide

Guide to the database layer as it exists in the Go Starter today: a generic,
business-agnostic storage abstraction over PostgreSQL (pgx/v5 + scany), wired
into the application lifecycle in `internal/app/app.go`.

> **Scope note.** This document describes the code that ships in the repository
> right now. The SQL strategy (raw SQL through a generic wrapper), the
> forward-only embedded migrations, and the absence of a seeding mechanism have
> known limitations that are tracked as open items (M-5, M-7, M-8) in
> `docs/reviews/2026-07-02-architecture-review.md`. Where this guide notes a
> "known limitation," see that review for the intended direction. Do not treat
> the review's target state (sqlc, squirrel, goose) as implemented — it is not.

## Overview

The project uses PostgreSQL with a **generic storage abstraction layer**. The
storage layer is intentionally business-agnostic: it executes SQL, scans rows
into caller-supplied structs, and manages transactions and connection
lifecycle. It contains no domain models and no domain queries.

- **Storage interfaces** (`internal/storage/`) — `Client`, `Transaction`,
  `RowScanner`, shared error values, and migration-status types.
- **PostgreSQL implementation** (`internal/storage/postgres/`) — a `Provider`
  built on pgx/v5, created via `Open` with functional options.
- **Domain models** — plain structs (e.g. `internal/models/user.go`) with `db`
  tags, scanned by the storage layer.

There is currently **no domain repository/service layer** and no `internal/user`
or `internal/order` package. Callers use the generic `Client` directly.

## Architecture

```
┌─────────────────────────────────┐
│  Controllers (ogen handlers)    │  ← internal/controller
├─────────────────────────────────┤
│  storage.Client interface       │  ← internal/storage/interfaces.go
├─────────────────────────────────┤
│  postgres.Provider (pgx/v5)     │  ← internal/storage/postgres
├─────────────────────────────────┤
│  Database (PostgreSQL)          │
└─────────────────────────────────┘
```

The `storage.Client` is created in `App.Run` and passed into the controller via
`controller.Options.Store`. See "Application Wiring" below.

## Storage Interfaces

Defined in `internal/storage/interfaces.go`. Every method that talks to the
database takes a `context.Context` as its first argument.

```go
// RowScanner reads query results into Go types.
type RowScanner interface {
    ScanOne(dst interface{}) error // first row into dst
    ScanAll(dst interface{}) error // all rows into *[]T
}

// Client is the core database interface.
type Client interface {
    Query(ctx context.Context, sql string, args ...interface{}) (RowScanner, error)
    Execute(ctx context.Context, sql string, args ...interface{}) error

    BeginTx(ctx context.Context) (Transaction, error)
    WithTx(ctx context.Context, fn func(tx Transaction) error) error

    Ping(ctx context.Context) error
    Close() error

    Name() string
    Logger() *zap.Logger
}

// Transaction is a Client-scoped unit of work.
type Transaction interface {
    Query(ctx context.Context, sql string, args ...interface{}) (RowScanner, error)
    Execute(ctx context.Context, sql string, args ...interface{}) error
    InsertMany(ctx context.Context, queries map[string][]interface{}) error
    Commit() error
    Rollback() error
}
```

> `storage.Provider` exists but is **deprecated** — it is an alias for `Client`.
> Use `Client` in new code. (Note the naming overlap with the concrete
> `postgres.Provider` struct, tracked as m-13 in the review.)

### Shared errors

`internal/storage/errors.go` exposes sentinel errors for consistent handling
across backends: `ErrNotFound`, `ErrDuplicate`, `ErrConnection`,
`ErrTransaction`, `ErrInvalidQuery`, `ErrTimeout`. Compare with `errors.Is`.

## PostgreSQL Implementation

`internal/storage/postgres` implements the interfaces with pgx/v5. The concrete
type is `Provider`, constructed with `Open`:

```go
// func Open(ctx context.Context, options ...Option) (*Provider, error)
store, err := postgres.Open(
    ctx,
    postgres.WithURI(ctx, "postgres://user:pass@localhost:5432/dbname"),
    postgres.WithPingTimeout(15),        // seconds
    postgres.WithLogger(logger),         // *zap.Logger
)
if err != nil {
    return fmt.Errorf("postgres.Open(): %w", err)
}
```

`Open` does **not** verify connectivity — pgxpool connects lazily. Call
`store.Ping(ctx)` to fail fast on an unreachable database (this is what
`App.Run` does at startup).

### Connection options

Configure the connection with exactly one of these options:

| Option | Use |
| --- | --- |
| `WithURI(ctx, uri)` | Recommended. Opens a `pgxpool.Pool` from a connection string. |
| `WithPool(pool)` | Use an existing `*pgxpool.Pool`. |
| `WithConnection(conn)` | Use a single `*pgx.Conn` (no pooling). |
| `WithMockPool(mock)` | Testing only — a `pgxmock` pool. |

Additional options: `WithPingTimeout(seconds int)` (pass `0` to disable the
timeout), `WithLogger(*zap.Logger)`.

> **Known limitation (m-8).** `WithMockPool` lives in production
> `options.go`, which pulls `pgxmock` into `go.mod` as a direct dependency. The
> review suggests moving mock wiring into a `_test.go` / `storagetest` package.

### Executing queries

`Query` returns a `RowScanner`; scan into a struct (or slice of structs) whose
fields carry `db` tags. Scanning is handled by
[scany](https://github.com/georgysavva/scany).

```go
// Single row
scanner, err := store.Query(ctx,
    `SELECT id, email, first_name, last_name FROM users WHERE id = $1`, id)
if err != nil {
    return nil, fmt.Errorf("query user: %w", err)
}
var u models.User
if err := scanner.ScanOne(&u); err != nil {
    return nil, fmt.Errorf("scan user: %w", err)
}

// Many rows
scanner, err = store.Query(ctx,
    `SELECT id, email, first_name, last_name FROM users ORDER BY email`)
if err != nil {
    return nil, err
}
var users []*models.User
if err := scanner.ScanAll(&users); err != nil {
    return nil, err
}

// Writes
if err := store.Execute(ctx,
    `INSERT INTO users (email, first_name, last_name) VALUES ($1, $2, $3)`,
    email, first, last); err != nil {
    return fmt.Errorf("insert user: %w", err)
}
```

Always use positional placeholders (`$1`, `$2`, …) — pgx parameterizes them,
which prevents SQL injection. Never interpolate user input into the SQL string.

> **Known limitation (M-5).** All access goes through this generic
> string-based `Query`/`Execute` API. There is no compile-time query safety and
> no query builder in the repo (neither `sqlc` nor `squirrel` is present). The
> review's target is sqlc-typed queries owning the pool for most access, with
> squirrel for dynamic queries. Until that lands, hand-written SQL is the
> pattern.

### Domain models

Models are plain structs with `db` tags, defined outside the storage layer. The
only one that ships today:

```go
// internal/models/user.go
type User struct {
    ID           string `json:"-" db:"id"`
    Email        string `json:"-" db:"email"`
    PasswordHash string `json:"-" db:"password_hash"`
    FirstName    string `json:"-" db:"first_name"`
    LastName     string `json:"-" db:"last_name"`
}
```

## Transactions

Prefer `WithTx`, which begins a transaction, runs your function, and commits —
rolling back automatically on error. A rollback failure is wrapped around the
original error so both are visible.

```go
err := store.WithTx(ctx, func(tx storage.Transaction) error {
    if err := tx.Execute(ctx,
        `INSERT INTO users (email, first_name) VALUES ($1, $2)`,
        email, first); err != nil {
        return err // triggers rollback
    }

    scanner, err := tx.Query(ctx,
        `SELECT id FROM users WHERE email = $1`, email)
    if err != nil {
        return err
    }
    var created models.User
    return scanner.ScanOne(&created)
})
if err != nil {
    return fmt.Errorf("create user tx: %w", err)
}
```

Both `WithTx` and `BeginTx` derive the transaction context with the provider's
ping timeout, so that value doubles as the overall transaction deadline — a
long-running transaction is cancelled once it elapses.

For lower-level control use `BeginTx`, but you then own `Commit`/`Rollback`:

```go
tx, err := store.BeginTx(ctx)
if err != nil {
    return err
}
defer func() { _ = tx.Rollback() }() // no-op after a successful Commit

if err := tx.Execute(ctx, `...`, args...); err != nil {
    return err
}
return tx.Commit()
```

`Transaction.InsertMany(ctx, map[string][]interface{})` batches multiple
statements over a single `pgx.Batch` within the transaction.

## Migrations

Migrations are **embedded SQL, forward-only**, run through the provider — there
is no `migrate` CLI, no `migrations/` directory convention, and no `make
migrate-*` targets. You embed `.sql` files with `//go:embed` and hand the
`embed.FS` to `Provider.Migrate`:

```go
import "embed"

//go:embed migrations/*.sql
var migrationsFS embed.FS

// On startup, after Open + Ping:
if err := store.Migrate(migrationsFS); err != nil {
    return fmt.Errorf("migrate: %w", err)
}
```

`Migrate` walks the FS for `.sql` files, compares them against a `migrations`
table (keyed by file path), and executes any that have not run — all inside one
transaction. `MigrationStatus(ctx, fs)` reports `StatusNone` / `StatusPartial`
/ `StatusComplete`.

> **Not yet wired.** `App.Run` opens and pings the database but does **not**
> call `Migrate` today, and the repo ships no `migrations/` directory. Wiring
> auto-migration (or a `migrate` subcommand) is open item M-7.
>
> **Known limitations (M-7).** The current migrator has no down migrations, no
> checksums (editing an already-applied file is silently skipped), uses the
> full file path as the migration ID (renaming re-runs it), and hardcodes
> `context.Background()`. The review recommends adopting goose or
> golang-migrate, which support `embed.FS`, versioning, checksums, and down
> migrations.

### Seeding

There is **no seeding mechanism** in the repo (open item M-8). The review
suggests a `seed` subcommand or a dev-only `--seed` flag running idempotent
seed SQL/Go.

## Application Wiring

The storage client is created once in `App.Run` (`internal/app/app.go`) and
shared for the process lifetime:

```go
store, err := postgres.Open(
    ctx,
    postgres.WithURI(ctx, application.cfg.Database.URI),
    postgres.WithPingTimeout(application.cfg.Database.PingTimeout),
    postgres.WithLogger(application.logger),
)
if err != nil {
    return fmt.Errorf("postgres.Open(): %w", err)
}
if err := store.Ping(ctx); err != nil {
    _ = store.Close()
    return fmt.Errorf("store.Ping(): %w", err)
}
defer store.Close() // closed only after all components have drained

controllerOpts := controller.Options{
    ReadyState: application.ready,
    Store:      store, // storage.Client
    Validator:  validator,
}
```

The pool intentionally outlives the component errgroup so in-flight requests can
still reach the database while the webserver drains on shutdown.

### Configuration

Database settings live under `Database` in `pkg/config/config.go` and are bound
via Viper in `cmd/serve.go`:

| Field | Flag | Env-friendly key | Default |
| --- | --- | --- | --- |
| `URI` | `--database-uri` | `database.uri` | `""` |
| `PingTimeout` | `--database-ping-timeout` | `database.ping-timeout` | `15` |

Example: `--database-uri "postgres://user:pass@localhost:5432/dbname"`.

## Testing Database Code

The storage package is tested with [pgxmock](https://github.com/pashagolub/pgxmock)
via the `WithMockPool` option — **not** a hand-rolled testify mock. Pattern (see
`internal/storage/postgres/postgres_test.go` for the real tests):

```go
mock, err := pgxmock.NewPool()
require.NoError(t, err)
defer mock.Close()

store, err := postgres.Open(ctx,
    postgres.WithMockPool(mock),
    postgres.WithLogger(zaptest.NewLogger(t)),
)
require.NoError(t, err)

mock.ExpectPing()
require.NoError(t, store.Ping(ctx))
require.NoError(t, mock.ExpectationsWereMet())
```

For higher-level tests you can also implement the `storage.Client` interface
directly with an in-memory fake.

## Best Practices

1. **Always pass context.** Every `Client`/`Transaction` method takes `ctx`;
   propagate the request context so cancellation and timeouts work.
2. **Parameterize queries.** Use `$1`, `$2`, … placeholders. Never build SQL by
   concatenating user input.
3. **Map driver errors to sentinels.** Translate `pgx.ErrNoRows` to
   `storage.ErrNotFound` and unique-violation (`23505`) to
   `storage.ErrDuplicate` so callers can `errors.Is` regardless of backend:

   ```go
   if errors.Is(err, pgx.ErrNoRows) {
       return nil, storage.ErrNotFound
   }
   var pgErr *pgconn.PgError
   if errors.As(err, &pgErr) && pgErr.Code == "23505" {
       return nil, storage.ErrDuplicate
   }
   ```
4. **Prefer `WithTx`** for multi-statement units of work — it handles
   commit/rollback and wraps rollback failures around the original error.

## Related Documentation

- [project-context.md](project-context.md) — project overview
- [foundations.md](foundations.md) — architectural foundations
- [../guidelines/interface-pattern.md](../guidelines/interface-pattern.md) — interface patterns
- [../guidelines/testing-strategy.md](../guidelines/testing-strategy.md) — testing approaches
- [../reviews/2026-07-02-architecture-review.md](../reviews/2026-07-02-architecture-review.md) — open items referenced above
- [pgx documentation](https://github.com/jackc/pgx)
- [scany documentation](https://github.com/georgysavva/scany)

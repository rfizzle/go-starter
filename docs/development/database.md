# Database Patterns Guide

The starter uses one PostgreSQL access pattern: a caller-owned pgx pool,
sqlc-generated queries for fixed SQL, squirrel for dynamic SQL, and
repository-owned transaction boundaries.

## Architecture

```text
internal/app
  ├── postgres.Client       startup ping and embedded Goose migrations
  └── user.Repository       domain persistence
        ├── db.Queries      sqlc-generated fixed queries
        └── db.DBTX         pgxpool.Pool or pgx.Tx for dynamic queries
```

`internal/storage` is deliberately small. It contains the shared
`ErrNotFound` persistence outcome and migration-status values; it does not
expose a generic database client. Consumers define narrow behavioral
interfaces where they need a test seam. For example, the seeder accepts only
the two user operations it calls.

The application creates the concrete PostgreSQL pool in
`internal/app/app.go`. The pool is shared by migration infrastructure and the
user repository, remains alive while HTTP requests drain, and is closed by the
composition root.

## Connection lifecycle

Build a production pool with `postgres.NewPool`. `PoolConfig` applies explicit
limits and lifetimes on top of the URI-derived pgx defaults:

```go
pool, err := postgres.NewPool(ctx, cfg.Database.URI, postgres.PoolConfig{
    MaxConns:                 cfg.Database.MaxConns,
    MinConns:                 cfg.Database.MinConns,
    MaxConnLifetimeSeconds:   cfg.Database.MaxConnLifetime,
    MaxConnIdleTimeSeconds:   cfg.Database.MaxConnIdleTime,
    HealthCheckPeriodSeconds: cfg.Database.HealthCheckPeriod,
})
if err != nil {
    return fmt.Errorf("create pool: %w", err)
}
defer pool.Close()
```

pgx pools connect lazily. The application wraps the pool in a concrete
`postgres.Client` and pings before starting the HTTP listener:

```go
store, err := postgres.Open(
    ctx,
    postgres.WithPool(pool),
    postgres.WithPingTimeout(cfg.Database.PingTimeout),
    postgres.WithLogger(logger),
)
if err != nil {
    return fmt.Errorf("open postgres lifecycle client: %w", err)
}
if err := store.Ping(ctx); err != nil {
    return fmt.Errorf("ping postgres: %w", err)
}
```

`postgres.Client` is lifecycle and migration infrastructure, not a query
abstraction for domain code. Repositories receive the pool directly.

## Fixed queries with sqlc

Migration SQL under `internal/storage/postgres/migrations/` is the schema
source for sqlc. Fixed queries live under
`internal/storage/postgres/queries/`; `make generate` writes their Go
implementation to `internal/storage/postgres/db/`.

Generated files are never edited directly. Change a migration or query source,
then run:

```sh
make generate
make generate-check
```

The user repository constructs `db.Queries` over the pool:

```go
type Repository struct {
    pool    *pgxpool.Pool
    exec    db.DBTX
    queries *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
    return &Repository{
        pool:    pool,
        exec:    pool,
        queries: db.New(pool),
    }
}
```

Use sqlc for fixed statements so SQL arguments and result scanning stay
compile checked. Translate driver-specific outcomes at the repository edge:

```go
row, err := r.queries.GetUserByEmail(ctx, email)
if errors.Is(err, pgx.ErrNoRows) {
    return user.User{}, storage.ErrNotFound
}
if err != nil {
    return user.User{}, fmt.Errorf("get user by email: %w", err)
}
```

`storage.ErrNotFound` is the intentionally shared persistence contract. Other
database failures retain their wrapped pgx error so the layer that owns a
specific policy can inspect it with `errors.Is` or `errors.As`.

## Dynamic queries with squirrel

Use squirrel only when the statement shape varies at runtime. Pin builders to
PostgreSQL placeholders and allowlist any dynamic identifiers:

```go
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

builder := psql.
    Select("id", "email", "first_name", "last_name", "created_at", "updated_at").
    From("users").
    OrderBy("created_at DESC")

if filter.EmailLike != "" {
    builder = builder.Where(sq.ILike{"email": "%" + filter.EmailLike + "%"})
}

query, args, err := builder.ToSql()
```

Values must remain bound arguments. Do not concatenate user input into SQL.
When iterating raw pgx rows, always close the rows and check the terminal
iteration error:

```go
rows, err := r.exec.Query(ctx, query, args...)
if err != nil {
    return nil, fmt.Errorf("list users: %w", err)
}
defer rows.Close()

for rows.Next() {
    // Scan the explicit column list.
}
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("iterate users: %w", err)
}
```

## Transactions

Repository methods run against the generated `db.DBTX` interface. Both
`*pgxpool.Pool` and `pgx.Tx` satisfy that query contract, so a transaction
creates a scoped repository without duplicating query methods:

```go
func (r *Repository) WithinTx(
    ctx context.Context,
    fn func(*Repository) error,
) (err error) {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer func() {
        if rbErr := tx.Rollback(ctx); rbErr != nil &&
            !errors.Is(rbErr, pgx.ErrTxClosed) {
            err = errors.Join(err, fmt.Errorf("rollback tx: %w", rbErr))
        }
    }()

    txRepo := &Repository{
        pool:    r.pool,
        exec:    tx,
        queries: r.queries.WithTx(tx),
    }
    if err = fn(txRepo); err != nil {
        return err
    }
    if err = tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit tx: %w", err)
    }
    return nil
}
```

The callback must not commit or roll back the transaction itself. Keep the
transaction short and do not perform external network I/O inside it.

## Migrations and seeding

Goose applies the embedded, forward-only migrations through
`postgres.Client.Migrate`. `serve` and `seed` ping first and apply pending
migrations before using repositories. Add schema changes as new sequential
migration files; never rewrite a migration that may already have run.

`MigrationStatus` reports `storage.StatusNone`, `storage.StatusPartial`, or
`storage.StatusComplete`. Status lookup ensures Goose's version table exists,
so it is not a strictly read-only operation on a pristine database.

The `seed` command constructs `user.Repository` over the same pool and passes
it to `internal/seed` through a narrow `UserStore` interface. Seed records are
matched by natural key and skipped when already present.

## Testing

- Test repository mappings and query behavior in the owning domain package.
- Use pgxmock for query arguments, row scanning, failures, and transaction
  boundaries.
- Use a real disposable PostgreSQL instance for constraints, migrations,
  locking, and rollback semantics that a mock cannot prove.
- Use small hand-written fakes for consumer interfaces such as
  `seed.UserStore`; do not fake a broad database client.

Run generation before tests:

```sh
make generate
go test ./internal/storage/... ./internal/user ./internal/seed
make check
make race
```

## Related documentation

- [Project context](project-context.md)
- [Development foundations](foundations.md)
- [Interface patterns](../guidelines/interface-pattern.md)
- [Testing strategy](../guidelines/testing-strategy.md)
- [pgx documentation](https://github.com/jackc/pgx)
- [sqlc documentation](https://docs.sqlc.dev/)
- [squirrel documentation](https://github.com/Masterminds/squirrel)
- [Goose documentation](https://pressly.github.io/goose/)

# Database Patterns Guide

The starter uses one PostgreSQL access pattern: a caller-owned pgx pool,
sqlc-generated queries for fixed SQL, squirrel for dynamic SQL, and
repository-owned transaction boundaries.

## Architecture

```text
internal/app
  ├── platform/database/postgres          startup ping and optional locked migrations
  └── modules/user.Repository             domain persistence adapter
        ├── db.Queries      sqlc-generated fixed queries
        └── db.DBTX         pgxpool.Pool or pgx.Tx for dynamic queries
```

`internal/platform/database` contains technical migration-status contracts and
the PostgreSQL implementation. Business-facing user data, persistence inputs,
and `domain.ErrNotFound` live in `internal/domain`, so controllers and CLI
workflows do not depend on database infrastructure. There is no shared
repository interface: each consumer defines the narrow behavioral port it
needs (and tests against), and `*user.Repository` satisfies it structurally;
the seeder accepts only the two user operations it calls.

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

pgx pools connect lazily. The short-lived commands ping with a bounded timeout
before doing any work, and the application registers `pool.Ping` as a named
required-dependency check: startup verifies it once, and every readiness
request reevaluates it within the configured ping budget.

```go
if err := postgres.Ping(ctx, pool, time.Duration(cfg.Database.PingTimeout)*time.Second); err != nil {
    return fmt.Errorf("postgres.Ping(): %w", err)
}
```

`internal/platform/database/postgres` exposes lifecycle and migration
functions over the caller-owned pool (`Ping`, `Migrate`, `MigrationStatus`);
it is not a query abstraction for domain code. Repositories receive the pool
directly.

## Fixed queries with sqlc

Migration SQL under `internal/platform/database/postgres/migrations/` is the schema
source for sqlc. Fixed queries live under
`internal/platform/database/postgres/queries/`; `make generate` writes their Go
implementation to `internal/platform/database/postgres/db/`.

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
    return domain.User{}, domain.ErrNotFound
}
if err != nil {
    return domain.User{}, fmt.Errorf("get user by email: %w", err)
}
```

`domain.ErrNotFound` is the shared business-facing absence contract. Other
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

Goose applies the embedded, forward-only migrations through the context-aware
`postgres.Migrate`. Every application attempt holds a PostgreSQL
session advisory lock for the complete Goose `up` operation, so concurrent
commands or startup migrators serialize instead of racing.

Use the explicit one-shot commands for normal development and production
deployment:

```sh
go run . migrate status
go run . migrate up
go run . seed
```

`migrate up` is idempotent and applies only pending forward migrations.
`migrate status` reports `none`, `partial`, or `complete`. Both commands return
database and migration failures through Cobra's nonzero error path and stop
when their context is cancelled.

The `serve` command does not migrate by default. Set
`database.auto-migrate=true` (or
`GOSTARTER_DATABASE_AUTO_MIGRATE=true`) only when the deployment explicitly
chooses startup migration; the same advisory lock serializes replicas and any
failure returns before the HTTP listener is constructed. Prefer a deployment
job or init container running `migrate up` before application replicas start.
An init container runs once per pod and is safe because of locking, but a
single deployment job avoids redundant attempts and gives schema rollout an
independent success gate.

The `seed` command assumes migrations are already complete and is limited to
the development environment. Its user identity comes from explicit seed
configuration (`seed.email` and `seed.password`, both required); the password
is never printed. Add schema changes as new sequential migration files; never
rewrite a migration that may already have run.

`MigrationStatus` reports `database.StatusNone`, `database.StatusPartial`, or
`database.StatusComplete`. Status lookup ensures Goose's version table exists,
so it is not a strictly read-only operation on a pristine database.

The `seed` command constructs `user.Repository` over the same pool and passes
it to `internal/cli/seed` through a narrow `UserStore` interface. Seed records
are matched by natural key and skipped when already present; existing names
and password hashes are never overwritten.

## Testing

- Test repository mappings and query behavior in the owning module package.
- Use pgxmock (`github.com/pashagolub/pgxmock/v4`, added on first use) for
  query arguments, row scanning, failures, and transaction boundaries.
- Use a real disposable PostgreSQL instance for constraints, migrations,
  locking, and rollback semantics that a mock cannot prove.
- Use small hand-written fakes for consumer interfaces such as
  `seed.UserStore`; do not fake a broad database client.

Run generation before tests:

```sh
make generate
go test ./internal/platform/database/... ./internal/modules/user ./internal/cli/seed
make check
make race
```

## Related documentation

- [Project context](project-context.md)
- [Module and platform boundaries](../decisions.md#d1-module-and-platform-boundaries)
- [Engineering and testing policy](../../AGENTS.md)
- [pgx documentation](https://github.com/jackc/pgx)
- [sqlc documentation](https://docs.sqlc.dev/)
- [squirrel documentation](https://github.com/Masterminds/squirrel)
- [Goose documentation](https://pressly.github.io/goose/)

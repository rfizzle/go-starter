---
name: golang-database
description: Write, review, test, and troubleshoot Go database code for PostgreSQL using pgx/v5, pgxpool, sqlc, squirrel, Goose, parameterized SQL, transactions, migrations, and repository patterns. Use for repository methods, SQL queries, connection pools, pgxmock tests, integration tests, migration changes, and database-to-domain error translation. Follow the repository's module/platform boundaries and approved backend specs.
license: MIT
---

# Go Database

Use the repository's established PostgreSQL stack. Do not introduce `database/sql`,
sqlx, an ORM, another query builder, or a second migration system without explicit
architectural approval.

## Project stack

- PostgreSQL 16
- `github.com/jackc/pgx/v5` and `pgxpool`
- sqlc-generated fixed queries under `internal/platform/database/postgres/db`
- `github.com/Masterminds/squirrel` for genuinely dynamic query shapes
- Goose with sequential SQL migrations under `internal/platform/database/postgres/migrations`
- `internal/domain` business-facing models, repository contracts, and sentinel errors
- `github.com/pashagolub/pgxmock/v4` for repository unit tests

Read the relevant module spec, repository implementation, migration history, and
platform database code before changing behavior.

## Repository rules

1. Keep repository implementations in the owning business module.
2. Use sqlc's generated `db.DBTX` seam so fixed queries work with a pool, transaction,
   or mock. Keep pool ownership and lifecycle in the composition root.
3. Use explicit column lists and `db` tags. Never rely on `SELECT *`.
4. Pass `context.Context` to every query and transaction operation.
5. Parameterize every value. PostgreSQL placeholders are `$1`, `$2`, and so on.
6. Use sqlc for fixed statements. Build genuinely dynamic filters with a local
   PostgreSQL-configured squirrel builder. Allowlist identifiers; they cannot be bound.
7. Map generated database rows to `domain` values inside the repository. Persistence-only
   fields must not escape the adapter.
8. Translate `pgx.ErrNoRows` to the module's documented domain sentinel and preserve
   other error chains with `%w`.
9. Close rows and check iteration errors when using raw pgx row iterators.

```go
func (r *Repository) ByEmail(ctx context.Context, email string) (domain.User, error) {
    row, err := r.queries.GetUserByEmail(ctx, email)
    if errors.Is(err, pgx.ErrNoRows) {
        return domain.User{}, domain.ErrNotFound
    }
    if err != nil {
        return domain.User{}, fmt.Errorf("get user by email: %w", err)
    }
    return fromDB(row), nil
}
```

Match the real helper signatures in the repository; the example shows the intended
shape, not a substitute API.

## Transactions and locking

- Use a transaction when multiple statements must commit or roll back as one invariant.
- Build a transaction-scoped repository whose sqlc queries and dynamic executor both
  use the same `pgx.Tx`; do not duplicate transactional query methods.
- Prefer a single atomic `UPDATE ... RETURNING` when it expresses the invariant.
- Use `SELECT ... FOR UPDATE` only when a read-then-write sequence genuinely requires a
  row lock.
- Use `FOR UPDATE SKIP LOCKED` for competing job claimers and keep the claiming
  transaction short.
- Retry serialization/deadlock failures only around an idempotent whole transaction and
  with bounded, context-aware backoff.

Read [transactions](references/transactions.md) for isolation and lock choices.

## Migrations

Migration SQL is allowed when the user requested the schema change and the relevant
approved spec defines the behavior. Before writing it:

1. Inspect adjacent migrations and the current schema.
2. Use `make migration name=<name>` to choose the next sequential version.
3. Write both Goose `Up` and `Down` paths unless irreversibility is explicitly approved.
4. Consider existing row count, lock duration, defaults, backfills, indexes, and rollout
   compatibility.
5. Keep secrets and environment-specific data out of migrations.
6. Never edit a migration that may already have been applied; add a new migration.

For destructive, long-locking, or data-rewriting migrations, present the rollout and
rollback risk before implementation.

## Testing and verification

- Use pgxmock for query shape, arguments, scanning, errors, and transaction boundaries.
- Use integration tests for PostgreSQL semantics, constraints, locking, migrations, and
  behavior pgxmock cannot prove.
- Keep tests deterministic and independently runnable.
- Before any test target, run the repository-required generation step.

```bash
make generate
go test ./internal/platform/database/... ./internal/modules/...
make race
make check
```

Use focused `go test` commands only after `make generate`, and preserve the race detector
when the project target normally enables it.

Read [scanning](references/scanning.md), [testing](references/testing.md), and
[performance](references/performance.md) only when those concerns are part of the task.

## Review gate

- SQL is parameterized and identifiers are allowlisted.
- Context and errors propagate correctly.
- Not-found and constraint behavior match the module error contract.
- Transaction scope preserves the documented invariant without unnecessary locking.
- Rows/resources are closed and terminal iterator errors are checked.
- Migration and rollback paths are reviewed when schema changes are present.
- Tests cover success, no rows, database failure, and relevant constraint/concurrency
  behavior.

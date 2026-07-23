---
name: golang-database
description: Write, review, test, and troubleshoot Go database code for PostgreSQL using pgx/v5, pgxpool, scany/pgxscan, parameterized SQL, transactions, queue locking, migrations, and repository patterns. Use for repository methods, SQL queries, connection pools, pgxmock tests, integration tests, migration changes, and database error translation. Follow the repository's database abstractions and approved backend specs.
license: MIT
---

# Go Database

Use the repository's established PostgreSQL stack. Do not introduce `database/sql`,
sqlx, an ORM, another query builder, or a second migration system without explicit
architectural approval.

## Project stack

- PostgreSQL 16
- `github.com/jackc/pgx/v5` and `pgxpool`
- `github.com/georgysavva/scany/v2/pgxscan`
- `internal/platform/database` abstractions and error translation
- `github.com/pashagolub/pgxmock/v4` for repository unit tests
- `golang-migrate` and sequential reversible SQL files under `migrations/`

Read the relevant module spec, repository implementation, migration history, and
`internal/platform/database` contracts before changing code.

## Repository rules

1. Keep repository implementations in the owning business module.
2. Accept the narrow query interface established by `internal/platform/database` so the
   same repository works with a pool, transaction, or mock.
3. Use explicit column lists and `db` tags. Never rely on `SELECT *`.
4. Pass `context.Context` to every query and transaction operation.
5. Parameterize every value. PostgreSQL placeholders are `$1`, `$2`, and so on.
6. Build dynamic filters through the existing `WhereBuilder` or an equally parameterized
   local pattern. Allowlist identifiers such as sort columns; identifiers cannot be bound.
7. Translate pgx/constraint outcomes through the platform sentinel helpers. Preserve
   error chains with `%w`.
8. Close rows and check iteration errors when using row iterators. Prefer the platform
   scanning helpers for ordinary one/all-result queries.

```go
func (r *Repository) GetByID(
    ctx context.Context,
    id uuid.UUID,
) (*domain.Session, error) {
    const query = `
        SELECT id, status, started_at, ended_at
        FROM sessions
        WHERE id = $1
    `

    var session domain.Session
    if err := database.ScanOne(ctx, r.db, &session, query, id); err != nil {
        return nil, fmt.Errorf("sessions.GetByID: %w", err)
    }
    return &session, nil
}
```

Match the real helper signatures in the repository; the example shows the intended
shape, not a substitute API.

## Transactions and locking

- Use a transaction when multiple statements must commit or roll back as one invariant.
- Pass the transaction's query interface into the same repository code instead of
  duplicating transactional methods.
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
2. Choose the next sequential version.
3. Write both `up` and `down` paths unless irreversibility is explicitly approved.
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
make go-test
make go-test-integration
make lint
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

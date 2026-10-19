---
name: golang-database
description: How PostgreSQL code is changed in this template — sqlc for fixed queries, squirrel for dynamic ones, Goose sequential migrations, pgx/v5 transactions through the repository's WithinTx, and the Makefile targets that regenerate and test them. Use when adding or changing a query, a migration, a repository method, a transaction, or database-to-domain error translation under internal/modules or internal/platform/database. Not for general SQL or pgx usage (their own docs), package placement (go-project-layout) or test design (golang-testing).
license: MIT
metadata:
  author: rfizzle
  standard: agentskills.io
---

# Go Database

The stack is fixed: PostgreSQL 17, pgx/v5 with `pgxpool`, sqlc, squirrel and
Goose. Don't add `database/sql`, sqlx, an ORM, another query builder or a
second migration system without a decision recorded for it.

## Where things are

- Query sources: `internal/platform/database/postgres/queries/` (sqlc input).
- Generated code: `internal/platform/database/postgres/db/` (sqlc output,
  never edited; `make generate-check` fails on drift).
- Migrations: `internal/platform/database/postgres/migrations/`, sequential
  `00001_name.sql` files, embedded in the binary and applied by `migrate up`.
- The repository for a capability lives in its module
  (`internal/modules/user/repository.go`) and maps generated rows to
  `internal/domain` values. Only `modules/user` and `platform/database/postgres`
  may import the generated `db` package (`make architecture-check`).
- The pool is built in `internal/app` from `internal/platform/config`; modules
  never open their own. It is shared by migrations, readiness and the
  repositories, and is closed after the HTTP server drains.
- pgx pools connect lazily: the short-lived commands ping within
  `database.ping-timeout` before doing any work, and the server registers the
  ping as a readiness check. `internal/platform/database/postgres` exposes only
  lifecycle and migration functions over the caller-owned pool; it is not a
  query abstraction, and repositories receive the pool directly.

## Change a fixed query

1. Edit or add the query in the `queries/` directory.
2. `make generate` (runs sqlc and ogen).
3. Call the generated method through the repository's `queries`, map the row
   with the module's `fromDB`-style function, and translate `pgx.ErrNoRows` to
   the domain's not-found sentinel. Other failures keep their wrapped pgx
   error so the layer that owns a policy can inspect it with `errors.Is` or
   `errors.As`. Persistence-only fields (password hashes) never leave the
   adapter.
4. `make generate-check`, then the tests below.

## Change a dynamic query

Use squirrel only when the statement's shape varies at runtime. Use the
module's PostgreSQL-configured builder (`$N` placeholders), select an explicit
column list, bind every value and allowlist identifiers (they cannot be bound).
Execute through the repository's `exec`, the same `db.DBTX` sqlc uses, so it
works inside a transaction. See [scanning](references/scanning.md).

## Transactions

Use the repository's `WithinTx`: it builds a transaction-scoped copy whose sqlc
queries and squirrel executor share one `pgx.Tx`, commits on success and rolls
back on error. Don't call `Commit` or `Rollback` in the callback or keep the
scoped repository. Prefer one `UPDATE ... RETURNING` when it holds the
invariant. See [transactions](references/transactions.md) for lock choices.

## Add a migration

1. Read the adjacent migrations and the current schema.
2. `make migration name=<snake_case_name>`. Never bare `goose create`: its
   timestamped name sorts after the sequential ones.
3. Write both the Goose `Up` and `Down` sections unless the person agreed the
   change is irreversible.
4. Never edit a migration that may have been applied; add a new one.
5. Consider row count, lock duration, defaults and backfills; for a
   destructive or long-locking migration, state the rollout and rollback risk
   before writing it.
6. Keep secrets and environment-specific data out of migrations.

## Apply migrations and seed

```bash
go run . migrate status   # none, partial or complete
go run . migrate up       # idempotent; pending forward migrations only
go run . seed             # development environment only
```

- Every `migrate up` holds a PostgreSQL session advisory lock for the whole
  Goose run, so concurrent commands and replicas serialize instead of racing.
- `serve` migrates only when `database.auto-migrate` is true (off by default);
  a failure returns before the listener is built.
- `migrate status` creates Goose's version table if it is missing, so it is
  not read-only on a pristine database.
- `seed` expects a current schema and needs `seed.email` and `seed.password`.
  It matches records by natural key and skips existing ones; it never
  overwrites a name or password hash and never prints the password.

## Test and verify

- Mapping and pure logic: unit tests beside the code (`make test`).
- Query arguments, row scanning and transaction boundaries without a
  database: pgxmock (`github.com/pashagolub/pgxmock/v4`, added on first use).
- Consumers of a repository (the seeder's user store, a service port): a
  small hand-written fake of the consumer's interface, never a broad database
  client.
- SQL, constraints, locking and migrations: integration tests with the
  `postgrestest` harness behind the `integration` tag. See
  [testing](references/testing.md).

```bash
make generate
go test ./internal/platform/database/... ./internal/modules/user ./internal/cli/seed
go test -tags integration ./internal/modules/... ./internal/platform/database/...
make test-integration
make check
```

## Review gate

- Every value is bound; identifiers are allowlisted.
- Not-found and constraint errors match the module's documented contract.
- Transaction scope holds the invariant without needless locking.
- Generated code and applied migrations are untouched; the new migration has a
  `Down`.

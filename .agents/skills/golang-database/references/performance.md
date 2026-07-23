# PostgreSQL Performance

Optimize only after a representative query, trace, or profile identifies database work
as the bottleneck. Preserve correctness, cancellation, and the repository abstraction.

## Connection pool

`postgres.NewPool` constructs the configured `pgxpool.Pool`; `postgres.Open` owns startup
ping and migrations over that caller-owned pool. Pool sizes, idle time, and lifetime come
from `internal/platform/config`; modules must not create their own pools.

For platform diagnostics, inspect `pool.Stat()` and expose bounded metrics such
as acquired, idle, total, acquire count, and acquire duration. Register collectors with
the application's explicit registry; do not use global registration from package init.

Tune from observed concurrency and database capacity. A larger pool can increase lock
contention and overload PostgreSQL.

## Query diagnosis

1. Capture the exact parameterized query and representative parameter distribution.
2. Use `EXPLAIN (ANALYZE, BUFFERS)` in a safe non-production or approved diagnostic
   environment.
3. Check rows estimated versus returned, sequential scans, sorts, buffer reads, lock
   waits, and round trips.
4. Add or change an index only with a migration and a rollback/rollout plan.
5. Benchmark again with realistic data volume.

Avoid `SELECT *`, unbounded lists, N+1 queries, and per-row network round trips. Use
deterministic pagination and explicit columns.

## Batch work

- For moderate batches, use one parameterized multi-row statement within PostgreSQL's
  parameter limits.
- For large inserts, use pgx `CopyFrom` through an approved platform capability rather
  than importing a concrete pool into a business module.
- Chunk large updates/deletes to bound locks and transaction duration.
- Keep queue claims short with `FOR UPDATE SKIP LOCKED`; perform expensive work after the
  claim commits.

Any new batch or copy API is a consumer-side port and composition-root wiring decision,
not an excuse for a module to reach into `*pgxpool.Pool`.

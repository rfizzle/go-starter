# PostgreSQL Transactions and Locking

The composition root owns the pool. A repository that owns a multi-statement invariant
may expose `WithinTx` and build a transaction-scoped copy whose generated `db.DBTX`
dependencies use the same `pgx.Tx`.

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
        if rollbackErr := tx.Rollback(ctx); rollbackErr != nil &&
            !errors.Is(rollbackErr, pgx.ErrTxClosed) {
            err = errors.Join(err, fmt.Errorf("rollback tx: %w", rollbackErr))
        }
    }()

    scoped := &Repository{
        pool: r.pool,
        exec: tx,
        queries: r.queries.WithTx(tx),
    }
    if err = fn(scoped); err != nil {
        return err
    }
    if err = tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit tx: %w", err)
    }
    return nil
}
```

`WithinTx` commits on callback success and rolls back on an error. Do not call `Commit`
or `Rollback` inside the callback, retain the scoped repository, or nest transactions.

## Lock choices

- Prefer one atomic `UPDATE ... WHERE ... RETURNING` when it can enforce the invariant.
- Use `SELECT ... FOR UPDATE` for a necessary read-then-write decision on the same row.
- Lock rows in a stable order when multiple rows participate, reducing deadlock risk.
- Use `FOR UPDATE SKIP LOCKED` for competing queue claimers and keep the claiming
  transaction short.
- Never hold a database transaction open across external network or object-storage I/O.

## Isolation and retries

The current repository starts pgx transactions with PostgreSQL's default isolation. A
capability that needs another isolation level requires an explicit module contract
change; do not open an unrelated pool inside business logic.

Retry serialization or deadlock failures only around an idempotent whole transaction,
with bounded context-aware backoff. Never retry an unknown partial side effect.

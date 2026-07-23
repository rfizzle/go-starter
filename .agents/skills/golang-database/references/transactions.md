# PostgreSQL Transactions and Locking

Use the injected `database.Storage.WithTx` boundary. Repository methods accept
`database.Querier`, so the same implementation runs with the pool, a transaction, or
pgxmock.

```go
func (s *Service) Transfer(
    ctx context.Context,
    fromID uuid.UUID,
    toID uuid.UUID,
    amount int64,
) error {
    return s.db.WithTx(ctx, func(tx database.Querier) error {
        accounts := NewRepository(tx)
        if err := accounts.Debit(ctx, fromID, amount); err != nil {
            return fmt.Errorf("debiting account: %w", err)
        }
        if err := accounts.Credit(ctx, toID, amount); err != nil {
            return fmt.Errorf("crediting account: %w", err)
        }
        return nil
    })
}
```

`WithTx` commits on success and rolls back on an error or panic. Do not call `Commit` or
`Rollback` inside the callback, and do not nest `WithTx` calls.

## Lock choices

- Prefer one atomic `UPDATE ... WHERE ... RETURNING` when it can enforce the invariant.
- Use `SELECT ... FOR UPDATE` for a necessary read-then-write decision on the same row.
- Lock rows in a stable order when multiple rows participate, reducing deadlock risk.
- Use `FOR UPDATE SKIP LOCKED` for competing queue claimers and keep the claiming
  transaction short.
- Never hold a database transaction open across external network or object-storage I/O.

## Isolation and retries

The platform helper currently starts pgx transactions with the default PostgreSQL
isolation. A capability that needs another isolation level requires an explicit platform
contract change; do not bypass `Storage` inside a business module.

Retry serialization or deadlock failures only around an idempotent whole transaction,
with bounded context-aware backoff. Never retry an unknown partial side effect.

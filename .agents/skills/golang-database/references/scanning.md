# Scanning PostgreSQL Results

Use sqlc-generated row and parameter types for fixed statements. Keep the generated code
inside `internal/platform/database/postgres/db` and map rows to business-facing
`internal/domain` types in the owning module.

## Fixed queries

Change query source under `internal/platform/database/postgres/queries`, run
`make generate`, and call the generated method through `*db.Queries`.

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

The mapping function is the persistence boundary. Do not expose generated `db` types,
`pgtype` values, or persistence-only fields such as password hashes to controllers or
business orchestration.

## Dynamic queries

Use squirrel only when the statement shape varies. Keep every value bound, select an
explicit column list, and scan into a local or generated row type whose fields match that
list.

```go
rows, err := r.exec.Query(ctx, query, args...)
if err != nil {
    return nil, fmt.Errorf("list users: %w", err)
}
defer rows.Close()

for rows.Next() {
    var row db.User
    if err := rows.Scan(&row.ID, &row.Email, &row.FirstName, &row.LastName); err != nil {
        return nil, fmt.Errorf("scan user: %w", err)
    }
}
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("iterate users: %w", err)
}
```

Use pointers or pgtype values only when SQL `NULL` semantics require them. Translate to
plain domain values at the adapter boundary.

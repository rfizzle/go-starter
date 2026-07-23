# Scanning PostgreSQL Results

Use `db` tags with the repository's scany/pgxscan helpers. Keep column lists explicit and
match aliases to field tags.

```go
type User struct {
    ID        uuid.UUID  `db:"id"`
    Name      string     `db:"name"`
    Bio       *string    `db:"bio"`
    DeletedAt *time.Time `db:"deleted_at"`
}
```

## One row

`database.ScanOne` maps scany/pgx no-row results to `database.ErrNotFound`.

```go
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*User, error) {
    const query = `
        SELECT id, name, bio, deleted_at
        FROM users
        WHERE id = $1
    `

    var user User
    if err := database.ScanOne(ctx, r.db, &user, query, id); err != nil {
        return nil, fmt.Errorf("users.Get: %w", err)
    }
    return &user, nil
}
```

Use `errors.Is(err, database.ErrNotFound)` at the layer that translates persistence
outcomes into the module's documented domain error.

## Many rows

`database.ScanAll` returns an empty non-nil slice when no rows match.

```go
func (r *Repository) ListActive(ctx context.Context) ([]*User, error) {
    const query = `
        SELECT id, name, bio, deleted_at
        FROM users
        WHERE active = $1
        ORDER BY name, id
    `
    users, err := database.ScanAll[User](ctx, r.db, query, true)
    if err != nil {
        return nil, fmt.Errorf("users.ListActive: %w", err)
    }
    return users, nil
}
```

Use raw `pgx.Rows` only when streaming or custom row handling is necessary. In that case,
close rows, scan every selected column, and check `rows.Err()` after iteration.

## NULL and JSON

Use pointers when the application meaningfully distinguishes SQL `NULL` from the type's
zero value. Keep database and JSON concerns separate: `db` tags control scanning; `json`
tags control transport. Use pgtype values only when their richer status semantics are
actually required.

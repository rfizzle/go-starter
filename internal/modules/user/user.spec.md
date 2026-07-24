# `user` Module Spec

> Status: draft · Layer: modules · Owner: backend · Depends on: domain, platform/database/postgres/db, pgx

## 1. Purpose

The user module implements PostgreSQL persistence for the starter's example
user capability. It maps database rows to infrastructure-free domain values,
supports fixed and filtered queries, and owns the user transaction boundary.
It does not authenticate users, hash credentials, map HTTP transport types,
own the database pool, or run migrations. It exposes a password hash only
through `CredentialByEmail`, scoped to the `domain.Credential` projection
consumed by authentication callers — never through `domain.User`.

## 2. Contract

**Satisfies:** `domain.UserRepository` (compile-time checked with
`var _ domain.UserRepository = (*Repository)(nil)`).

**Constructor:**

```go
func NewRepository(pool *pgxpool.Pool) *Repository
```

**Repository methods:**

```go
func (r *Repository) ByID(ctx context.Context, id string) (domain.User, error)
func (r *Repository) ByEmail(ctx context.Context, email string) (domain.User, error)
func (r *Repository) CredentialByEmail(ctx context.Context, email string) (domain.Credential, error)
func (r *Repository) Create(ctx context.Context, params domain.CreateUserParams) (domain.User, error)
func (r *Repository) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, error)
func (r *Repository) WithinTx(ctx context.Context, fn func(*Repository) error) error
```

The callback passed to `WithinTx` receives a transaction-scoped repository. It
must not commit, roll back, retain, or use that repository after the callback.

## 3. Data Model

Owns persistence access to the `users` table. Reads and returns `domain.User`;
accepts `domain.CreateUserParams` and `domain.UserFilter`. Password hashes are
write-only at the module boundary.

## 4. Behaviors & Invariants

1. **I1 — Absence:** `ByID` and `ByEmail` return `domain.ErrNotFound` when no
   user matches.
2. **I2 — Identifier validation:** `ByID` rejects a malformed UUID before
   issuing a query and preserves the parse error as wrapped context.
3. **I3 — Credential isolation:** no returned `domain.User` contains or can
   represent a password hash.
4. **I4 — Creation ownership:** `Create` persists the already-hashed
   `PasswordHash` it receives; the module does not hash plaintext credentials.
5. **I5 — Listing:** `List` orders results by newest creation time, applies an
   optional parameterized case-insensitive email pattern, and treats a zero
   limit as no explicit limit.
6. **I6 — Transaction outcome:** `WithinTx` commits only after a nil callback
   result, rolls back on callback failure, and preserves both the primary and
   rollback errors when both occur.
7. **I7 — Resource ownership:** the caller owns the pgx pool; repositories do
   not close it.
8. **I8 — Credential exposure boundary:** `CredentialByEmail` is the only
   method that returns a password hash, and it does so only as
   `domain.Credential`, never `domain.User`.

## 5. Error Contract

- Missing users return `domain.ErrNotFound`.
- Invalid IDs and database failures are wrapped with operation context and
  remain inspectable with `errors.Is` or `errors.As`.
- A callback error from `WithinTx` is returned without translation. Begin,
  rollback, and commit failures retain their original cause through wrapping.

## 6. Collaborators

| Dependency | Kind | Used for |
| --- | --- | --- |
| `internal/domain` | contract | user data, inputs, filter, repository interface, and not-found error |
| `internal/platform/database/postgres/db` | infrastructure adapter | generated fixed queries and row types |
| `pgxpool.Pool` / `pgx.Tx` | infrastructure adapter | query execution and transaction scoping |

`domain` and `platform` remain independent leaves. This module imports no
generated transport and is constructed in a composition root.

## 7. Testing Scenarios

### Unit

- [ ] I1: translate no-row outcomes for ID and email lookups.
- [x] I2: reject malformed UUIDs and preserve valid UUID round trips.
- [x] I3: database-to-domain mapping drops password hashes.
- [x] I8: `CredentialByEmail` retains the password hash; `ByEmail`/`fromDB` still drop it.
- [ ] I4: map creation parameters without changing the supplied hash.
- [ ] I5: bind filter values and preserve deterministic ordering.
- [ ] I6: cover commit, callback rollback, commit failure, and joined rollback failure.
- [ ] I7: repository operations never close the caller's pool.

### Integration

- [ ] Create and read a user against a disposable PostgreSQL database.
- [ ] Verify filtered listing and transaction commit/rollback against PostgreSQL.
- [ ] I1/I8: `CredentialByEmail` returns `domain.ErrNotFound` for an unknown email and the stored hash for a known one.

## 8. Open Questions

None.

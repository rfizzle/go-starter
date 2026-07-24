# `credential` Module Spec

> Status: approved · Layer: modules · Owner: backend · Depends on: domain, golang.org/x/crypto/bcrypt

## 1. Purpose

The credential module verifies a plaintext password against a user's stored
bcrypt hash and reports a single, timing-equalized outcome. It does not hash
passwords for storage, issue tokens or cookies, map HTTP transport, or read
any user data beyond the authentication-only projection it consumes.

## 2. Contract

**Constructor:**

```go
func NewService(reader credentialReader, hashCost int) (*Service, error)
```

**Methods:**

```go
func (s *Service) Authenticate(ctx context.Context, email, password string) (Authenticated, error)
```

`credentialReader` is a package-private port:

```go
type credentialReader interface {
    CredentialByEmail(ctx context.Context, email string) (domain.Credential, error)
}
```

It is satisfied structurally by `internal/modules/user.Repository`; this
package does not import that module.

## 3. Data Model

None — pure, in-memory verification. Reads `domain.Credential` through the
port above; returns `Authenticated{UserID, Email}`, which never carries a
password hash.

## 4. Behaviors & Invariants

1. **I1 — Bounded error:** `Authenticate` returns `ErrInvalidCredentials` for
   an unknown email, an incorrect password, or a stored hash that is malformed
   or does not use the configured cost; no other error type distinguishes
   these cases.
2. **I2 — Equalized verification:** every completed credential check performs
   exactly one bcrypt comparison at the constructor's configured cost.
   Missing accounts and unusable stored hashes use a fixed, non-secret dummy
   hash at that same cost.
3. **I3 — Distinct failure propagation:** a reader error other than
   `domain.ErrNotFound` — including context cancellation reported by the
   reader itself — is wrapped and returned as-is, never converted to
   `ErrInvalidCredentials`.
4. **I4 — Cancellation cannot be outrun:** `Authenticate` checks
   `ctx.Err()` both before reading and after the bcrypt comparison, so a
   reader that ignores cancellation and returns a stale, otherwise-successful
   result still yields a cancellation error, never a successful
   `Authenticated`.
5. **I5 — No hash exposure:** `Authenticated` and every error returned by
   `Authenticate` are free of the password hash.
6. **I6 — Construction guard:** `NewService` returns an error when `reader`
   is nil or `hashCost` is outside bcrypt's supported range.

## 5. Error Contract

- `ErrInvalidCredentials` is the sentinel for "the supplied credentials are
  wrong," covering unknown email, wrong password, malformed hash, and
  unexpected hash cost alike.
- Any other error — context cancellation/deadline (whether observed directly
  or reported by the reader) or a repository failure — is wrapped with
  `fmt.Errorf("credential: %w", err)` or
  `fmt.Errorf("credential: read credential: %w", err)` and remains
  inspectable with `errors.Is`/`errors.As`; it is never mapped to
  `ErrInvalidCredentials`.

## 6. Collaborators

| Dependency | Kind | Used for |
| --- | --- | --- |
| `internal/domain` | contract | `Credential` value type and `ErrNotFound` sentinel |
| `golang.org/x/crypto/bcrypt` | library | constant-behavior hash comparison |
| `internal/modules/user` (via injected `credentialReader`) | sibling, structural | supplies `CredentialByEmail` in this starter; never imported directly |

`domain` remains an independent leaf. This module imports no generated
transport, no controller, and no concrete storage package.

## 7. Testing Scenarios

### Unit

- [x] I1/I5: correct credentials authenticate and return no password hash.
- [x] I1: wrong password returns `ErrInvalidCredentials`.
- [x] I1/I2: unknown email returns `ErrInvalidCredentials` through the equalized comparison path.
- [x] I1/I2: malformed and unexpected-cost hashes use the equalized comparison path and return `ErrInvalidCredentials`.
- [x] I3: a non-not-found reader error is wrapped and distinct from `ErrInvalidCredentials`.
- [x] I3: a reader-reported cancellation error is wrapped and distinct from `ErrInvalidCredentials`.
- [x] I4: a context cancelled before the read short-circuits without calling the reader.
- [x] I4: a context cancelled during the read (reader still returns valid credentials) still denies success.
- [x] I6: invalid constructor dependencies return an error.

## 8. Open Questions

None.

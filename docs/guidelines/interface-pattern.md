# Consumer-Owned Interface Pattern

This project uses interfaces at real behavioral seams, not as mirrors of every
concrete implementation.

## Decision rule

Define an interface when at least one of these is true:

- a consumer needs a narrower contract than the provider exports;
- more than one real implementation must satisfy shared behavior; or
- a stable dependency boundary needs substitution in tests.

Otherwise keep the concrete type. Constructors normally return concrete types.
Interfaces belong beside the consumer unless several packages genuinely share
the same business contract.

## Placement

| Contract | Owner |
| --- | --- |
| Shared business data, repository behavior, and business errors | `internal/domain/` |
| Behavior needed by one business module | that module |
| Behavior needed by a controller | `internal/controller/` |
| Behavior needed by CLI orchestration | `internal/cli/` or its subpackage |
| Technical implementation | matching `internal/platform/` package |

Domain contracts must remain free of platform, module, controller, and
generated transport imports. Platform packages must not import domain or
modules.

## Example: controller consumer

The controller needs only an email lookup, not the whole repository:

```go
// internal/controller/controller.go
type UserReader interface {
    ByEmail(ctx context.Context, email string) (domain.User, error)
}

type Options struct {
    Users UserReader
}
```

`*user.Repository` satisfies this contract structurally. The controller can use
a small fake in tests without importing the concrete module or a database.

## Example: module-local dependency

Credential verification owns a private reader contract for the authentication
projection it consumes:

```go
// internal/modules/credential/service.go
type credentialReader interface {
    CredentialByEmail(ctx context.Context, email string) (domain.Credential, error)
}

type Service struct {
    reader credentialReader
}

func NewService(reader credentialReader, bcryptCost int) (*Service, error) {
    // validate current constructor inputs
}
```

The interface remains private because no other package needs to name it. The
user repository satisfies it without coupling the credential module to
PostgreSQL or generated SQL.

## Example: shared domain repository

`domain.UserRepository` is shared by the owning user adapter and CLI seed
orchestration. It exposes business values and errors, not pgx or sqlc types:

```go
type UserRepository interface {
    ByID(context.Context, string) (User, error)
    ByEmail(context.Context, string) (User, error)
    Create(context.Context, CreateUserParams) (User, error)
    List(context.Context, UserFilter) ([]User, error)
}
```

The concrete implementation lives in `internal/modules/user` and maps database
values at its edge. The PostgreSQL platform package does not expose a broad
application database interface.

## Constructors and assertions

- Return `*ConcreteType` from constructors.
- Accept the smallest dependency contract the implementation consumes.
- Validate constructor inputs only when construction can meaningfully fail.
- Add a compile-time assertion when a concrete implementation is intended to
  satisfy a named cross-package contract:

```go
var _ domain.UserRepository = (*Repository)(nil)
```

Do not create an interface solely so a constructor can return it.

## Contract tests

Use one reusable contract suite only when multiple maintained implementations
must share meaningful behavior. Test observable inputs, outputs, errors,
bounds, and ownership rules. Keep provider-specific behavior in
provider-specific tests.

A single implementation generally needs direct package tests, not a
conformance framework.

## Review checklist

- Is the interface required by a current consumer or real implementation set?
- Is it located with the consumer or in a genuinely neutral business contract?
- Does it expose only behavior the consumer uses?
- Are parameters and results free of provider internals?
- Does the constructor return a concrete type?
- Do errors and cancellation cross the boundary intentionally?
- Does a compile-time assertion protect an intended named contract?
- Would deleting the interface make the design simpler without losing a real
  seam? If yes, delete it.

## Anti-patterns

- Provider-owned interfaces that mirror every method on a concrete type
- `ports/`, `adapters/`, or `interfaces/` directories without multiple real
  concepts to organize
- Generic storage interfaces added for hypothetical database swaps
- Interfaces created only for mocking trivial code
- Transport or pgx/sqlc types leaking into business contracts
- One-method wrappers that add naming indirection but no policy boundary

See the [architecture decision](../architecture/module-platform-boundaries.md)
for dependency direction and the
[current architecture map](../architecture/current-state.md) for the concrete
packages.

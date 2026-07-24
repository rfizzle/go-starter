# Development Guidelines

These guides define the maintained Go Starter engineering rules. Repository
policy in `AGENTS.md` takes precedence when guidance overlaps.

## Guides

- [Go coding standards](go-standards.md) — naming, context, errors,
  concurrency, documentation, and measured performance work
- [Consumer-owned interfaces](interface-pattern.md) — when an interface is
  justified, where it belongs, and how implementations are wired
- [Testing strategy](testing-strategy.md) — risk-based test selection,
  integration boundaries, reliability, and coverage policy

Package placement and import direction are defined separately by the accepted
[module/platform boundary decision](../architecture/module-platform-boundaries.md).

## Priority

When requirements compete:

1. correctness and safety;
2. approved acceptance criteria;
3. simplicity and clarity;
4. maintainability; and
5. performance, extensibility, and reuse only when required or measured.

Do not add abstraction, configurability, concurrency, caching, retries, queues,
or dependencies for hypothetical future use.

## Quick review checklist

- Is every behavior traceable to the request or approved specification?
- Is the change in the owning module, platform package, edge, or composition
  root?
- Are generated ogen and sqlc files unchanged by hand?
- Are inputs, resource use, time, pagination, and concurrency bounded where
  relevant?
- Does cancellation reach blocking operations?
- Are returned and logged errors useful without exposing secrets?
- Does each new test protect a contract, boundary, regression, or named risk?
- Is the selected test boundary the narrowest one that can catch the defect?
- Does the change avoid speculative interfaces and helper layers?

## Required validation

Run focused checks while developing. Before completion, use the repository
targets appropriate to the changed risk:

```sh
make generate-check
make architecture-check
make docs-check
make check
```

Add `make race` for concurrency, lifecycle, middleware, storage, or test
infrastructure changes. Add `make test-integration` or `make test-all` when the
important behavior belongs to PostgreSQL or another real external boundary.

`make cover` enforces the maintained-code regression floor. Coverage does not
replace contract- and risk-based test selection.

## Current examples

- `internal/controller.UserReader` is a consumer-owned read seam.
- `internal/modules/credential` owns a private credential-reader interface.
- `internal/domain.UserRepository` is a shared business persistence contract.
- `internal/modules/user.Repository` is the concrete PostgreSQL-backed adapter.
- `internal/app` constructs and injects concrete dependencies.

Use these implemented examples and the
[current architecture map](../architecture/current-state.md), not generic
package-layout recipes.

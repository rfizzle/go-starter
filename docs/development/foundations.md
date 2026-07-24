# Go Project Development Foundations

These are the project-specific foundations for working in Go Starter. The
repository requires Go 1.26.5, as declared in `go.mod`.

## Package ownership

The repository uses a module/platform layout, not a generic directory template:

```text
api/                         OpenAPI source
cmd/                         Cobra commands and process concerns
internal/
  api/v1/                    generated ogen transport
  app/                       composition root and runtime lifecycle
  auth/                      JWT and generated security edge
  cli/                       command support
  controller/                transport/domain mapping
  domain/                    shared infrastructure-free business contracts
  modules/<capability>/      business behavior and adapters
  platform/<capability>/     technical infrastructure
  templateinit/              standalone initializer implementation
tools/initialize/            initializer entrypoint
web/                         embedded static placeholder
```

Apply the complete
[module/platform boundary decision](../architecture/module-platform-boundaries.md):

- domain imports no other application package;
- platform imports neither domain nor modules;
- modules never import generated transport;
- generated transport stays at edge packages and the composition root;
- controllers map types and delegate behavior; and
- catch-all packages such as `utils`, `helpers`, `common`, or `shared` are not
  allowed.

## Contract-first package design

Document a package's purpose, non-goals, inputs, outputs, bounds, errors,
concurrency guarantees, and resource ownership before adding abstraction.

An interface is justified by a real substitution boundary, multiple
implementations, or a consumer that needs a narrower seam. Prefer
consumer-defined interfaces and concrete constructor returns. One natural
implementation stays in its owning package; multiple real variants may use
clearly named subpackages.

See [interface-pattern.md](../guidelines/interface-pattern.md) for examples from
the current codebase.

## Context, errors, and resources

- Accept `context.Context` first for blocking, I/O, or cancellable work.
- Propagate the caller's context rather than replacing it.
- Derive bounded timeouts at system boundaries.
- Return and wrap errors with useful operation context.
- Translate infrastructure errors to the smallest business-facing contract at
  the adapter boundary.
- Do not include secrets, credentials, database URIs, or raw internal failures
  in public responses.
- Pair every acquired resource or background goroutine with explicit cleanup.

Do not add retries, caching, queues, fallback behavior, or concurrency without
a documented failure mode and bound.

## Generated boundaries

`api/openapi.yml` is the HTTP contract. `make generate` writes ogen output to
`internal/api/v1/`.

Migrations and SQL query files under
`internal/platform/database/postgres/` are the PostgreSQL sources.
`make generate` writes sqlc output to the nested `db/` package.

Generated files are reviewable artifacts but never hand-edited. Use
`make generate-check` to prove committed output matches its sources.

## Test selection

Tests protect useful behavior, not test counts. Add a test for an acceptance
criterion, public contract, meaningful boundary, significant failure mode,
regression, or security/concurrency/persistence invariant.

Choose the narrowest stable test that can catch the defect:

| Risk | Test |
| --- | --- |
| isolated decision or validation | unit/package test |
| HTTP mapping or status contract | controller or HTTP integration test |
| SQL, migration, constraint, lock, scan, or transaction | real PostgreSQL integration test |
| lifecycle across CLI, app, listener, and database | focused end-to-end test |
| shared concurrent state | deterministic test plus race detector |

Do not test generated code, trivial assignments, third-party behavior, or
private call order. See [testing-strategy.md](../guidelines/testing-strategy.md).

## Development loop

```sh
make generate       # only when generator inputs changed
make fmt            # when Go or YAML changed
make test-unit      # focused/default package behavior
make check          # repository fast gates
make race           # concurrency/lifecycle risk
make test-integration # real external-boundary risk
```

`make check` includes architecture, documentation, generation, formatting,
vet, lint, and unit checks. `make cover` reports total and maintained-code
coverage; its threshold is a regression floor, not a reason to add low-value
tests.

## Simplicity rules

- Implement only approved behavior and explicit acceptance criteria.
- Prefer the standard library and existing project patterns.
- Keep code explicit and local until a stable concept warrants extraction.
- Do not add a package, dependency, interface, option, or extension point for a
  hypothetical future consumer.
- Preserve small local duplication when an abstraction would hide behavior.
- Explain constraints and reasons in comments; do not narrate obvious code.

These rules are enforced by review and, where practical, by
`make architecture-check` and `make docs-check`.

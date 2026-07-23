---
name: go-project-layout
description: "Organize a growing Go service into business modules, technical platform packages, a central domain contract leaf, a quarantined generated transport/controller edge, and an explicit composition root. Use when deciding where a file or package belongs, designing module ports and test seams, reviewing import direction, splitting or merging packages, handling shared domain models, or avoiding utils/common grab bags. Not for individual identifier naming or in-package patterns."
license: MIT
---


# Go Project Layout

How to organize a growing Go service so that infrastructure and business code stay separable, dependencies point one way, and a developer can tell at a glance where any file belongs.

> "A little copying is better than a little dependency." — Go Proverbs
>
> "Clear is better than clever." — Go Proverbs

To deliberately ignore a rule, add a short comment at the package or file explaining why.

## When this applies

- A Go service whose `internal/` has grown past ~15 packages, mixing plumbing (config,
  database, metrics) with business capabilities.
- Creating a **new package or file** and needing to decide where it lives.
- Someone proposes a heavyweight rewrite (Clean / Hexagonal / "feature modules") — evaluate it against what's already idiomatic before adopting.

This skill is **not** about naming identifiers (see `golang-naming`) or about in-package design patterns (see `golang-design-patterns`).

## The target shape

Two buckets under `internal/`, plus a small set of packages that stay top-level because they are neither pure infrastructure nor a single business domain.

```
cmd/<app>/main.go            # thin entrypoint → cli.Execute()

internal/
  api/<gen>/                 # generated transport code (ogen, oapi-codegen) — isolated, DO NOT EDIT
  platform/                  # technical infrastructure (no business meaning)
    config/  database/  metrics/  middleware/  health/  web/  worker/
    database/sqlbuilder/     # DB utilities nest under the capability they serve
  modules/                   # business domains (the product surface)
    accounts/  orders/  billing/  reporting/  ...
  domain/                    # shared models + repository INTERFACES (central, flat)
  controller/                # the single generated-Handler implementation (transport ↔ domain mapping)
  auth/                      # cross-cutting (RBAC/OIDC) — touched only by edge layers
  app/                       # composition root: Start/Stop/Ready, wires everything
  cli/                       # cobra commands
```

`platform/` and `modules/` are **grouping directories only** — Go cares about import paths, not nesting depth. The payoff is human legibility: opening `internal/` shows infra-vs-product immediately, and the business domains become the visible surface.

## Quick reference — where does a new file go?

| The file is… | It belongs in… |
| --- | --- |
| Config loading, secrets, env | `platform/config/` |
| DB pool, transactions, migrations, query builders | `platform/database/` (+ nested `sqlbuilder/`) |
| Prometheus, tracing, log setup | `platform/metrics/` (or `platform/observability/`) |
| HTTP middleware (cors, logging, recovery, rate limit) | `platform/middleware/` |
| HTTP server construction + routing | `platform/web/` |
| Background queue / job worker infrastructure | `platform/worker/` |
| A business entity's model + its repository impl + its business logic | `modules/<domain>/` |
| A model struct or repository **interface** shared across domains | `domain/` |
| Mapping a generated request/response ↔ domain type | `controller/` |
| Auth / RBAC / token scopes | `auth/` |
| Dependency wiring, lifecycle | `app/` |
| A cobra command or dev tooling (seed, fixtures) | `cli/` |
| "It's kind of generic and used everywhere" | **STOP.** It's a `domain` type or a `platform` capability. Never `utils/`. |

## The rules

### Rule 1 — `domain` and `platform` are independent leaves

- `modules/*` may import `domain` contracts and selected `platform` abstractions.
- `domain` imports **nothing** from `modules`, `controller`, or `platform` I/O. It is a leaf.
- `platform/*` imports **nothing** from `modules` or `domain`.
- `controller` imports `domain`, `api/<gen>`, and edge helpers — never a `modules` package's internals.

This is a fan-out from modules to two leaves, not a chain where `domain` imports
`platform`:

```text
modules/* ──→ domain
        └──→ platform/*

domain      (leaf)
platform/*  (leaf with respect to business code)
```

If you feel the urge to import "upward" (a `platform` package reaching into a `module`, or `domain` importing a repository implementation), the type you need belongs one layer down. Move it to `domain`.

```
✓ modules/orders → domain.OrderRepository (interface)     // depend on the abstraction
△ modules/orders → modules/auditlog                       // only an established, reviewed seam
✗ domain → platform/database                               // leaf must not touch I/O
```

### Rule 2 — Keep a central `domain` package; do NOT dissolve it into modules

`domain` holds shared model structs and **repository interfaces** (the ports). Concrete Postgres implementations live in each `modules/<domain>/repository.go`.

This lets two domains reference shared types without an import cycle, and it mirrors the
mainstream idiomatic Go feature layout. Dissolving `domain` into per-module files can
create cycles as soon as two business capabilities need each other's types.

```go
// internal/domain/order.go — the contract layer, no I/O
type Order struct { ID uuid.UUID; AccountID uuid.UUID; Status string /* ... */ }

type OrderRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*Order, error)
}
```

### Rule 3 — Quarantine generated transport at the edge

Generated API code (ogen, oapi-codegen) typically exposes a **single large `Handler` interface**. Go interface satisfaction is all-or-nothing on one type, so that interface has exactly one implementer — keep it that way.

- Only approved edge packages (`controller/`, `auth/`, and `platform/web/`) plus the
  composition root import `internal/api/<gen>`.
- No `modules/*` package imports generated transport types.
- The controller's job is **mapping**: generated request → domain call → generated response. That mapping code is verbose and generator-specific; concentrating it means regenerating the API has a blast radius of one package.

Splitting the handler per-feature (`modules/alerts/handler.go`, etc.) spreads the generated dependency across every domain and buys nothing — reject it.

### Rule 4 — Wire dependencies manually at the composition root

Construct concrete types once in `app/` and inject them through explicit constructors.
Use direct dependency parameters by default. Keep functional options where the package
already establishes them or where optional configuration genuinely benefits from them.
Do not add `wire`/`dig` merely to hide visible wiring.

```go
ctrl := controller.New(
    controller.WithLogger(logger),
    controller.WithOrderRepository(orders.NewRepository(db)),
    controller.WithAccountRepository(accounts.NewRepository(db)),
)
```

This keeps DI visible, compile-checked, and greppable. Constructors follow `New<Type>(deps) *Type`.

### Rule 5 — Prefer data + functions over "rich" self-mutating models

Idiomatic Go keeps `domain` structs as data with validating constructors (`NewPagination`,
`NewSort`) and puts behavior in repositories and, where warranted, services. Be cautious
with self-mutating domain methods when an atomic repository operation is required; a
read-modify-write sequence can reintroduce races.

"Anemic domain model" is only an anti-pattern under full DDD with rich aggregates. For CRUD-over-Postgres services, data structs + interfaces are correct, not a smell.

### Rule 6 — Add a `Service` layer only where there is real orchestration

Most handlers are thin list/get mappings, so a service layer there is bloat. Introduce
`modules/<domain>/service.go` **only** for operations that orchestrate across multiple
repositories, such as updating an order, reserving inventory, and writing an audit event.

- The service lives **inside the module**, depends on `domain` interfaces, and has **zero** dependency on generated transport or HTTP.
- The controller then calls `service.Acknowledge(...)` and only maps the result.

### Rule 7 — No `utils` / `helpers` / `common` / `shared` packages

These grab-bags accumulate unrelated code and tangle the dependency graph. Name a package after what it *is* (`sqlbuilder`, not `dbutil`). Anything "truly shared" is either a `domain` type or a `platform` capability — file it there.

### Rule 8 — Package name = directory name, one clear noun

Short, lowercase, no stutter (`alerts.Repository`, not `alerts.AlertRepository`). If you cannot name a package in one clear noun, its scope is probably wrong — split or merge it.

### Rule 9 — Tests live beside the code they test

`*_test.go` stays in-package (or `_test` package for black-box). Layout regrouping never separates a test from its subject.

### Rule 10 — A module defines its own consumer-side ports; the concrete adapter is injected

Rule 1 keeps cross-package dependencies acyclic. Rule 10 is the in-module corollary for
business orchestration: it must not reach out and construct an infrastructure concrete
(`*storage.ObjectStore`, `*database.PgxStore`, or a third-party SDK client). It declares a small local
interface describing only the behavior it needs, and `app/` injects a concrete that
structurally satisfies it. A repository adapter that implements persistence may import
the established `platform/database` query abstraction; the business logic above it still
depends on a port.

This is dependency inversion via Go's structural typing. It is **not** extra ceremony on
top of Rule 2—Rule 2's `domain` repository interfaces are the shared, cross-domain ports;
Rule 10 covers module-private seams for storage or third-party services that do not belong
in `domain`.

For example, a reporting module may need object storage, a renderer, and a repository
without importing any concrete implementation. It declares only the behavior it consumes:

```go
// reporting/handler.go — ports defined by the consumer, satisfied structurally
type objectStore interface { GetObject(ctx context.Context, key string) ([]byte, error) }
type reportStore interface {
    Create(ctx context.Context, report *domain.Report) error
    ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Report, error)
}
type accountGetter interface {
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
}
```

Rules of thumb:

- **Ports are unexported and live with the consumer**, named for the behavior (`recordingStore`, `budgeter`), not the implementer. Keep them 1–3 methods (see `golang-structs-interfaces`).
- **Narrow at the seam, even over a type you own.** A worker can depend on a local
  `claimer` interface instead of a complete repository, so a fake drives unit tests
  without a database.
- **The concrete never appears in the module** — only in the composition root, where the
  constructor receives concrete implementations. The module cannot tell whether it is
  using a cloud service or a fake.
- **The payoff is testability.** Every seam is a place a unit test can inject a fake, so
  module tests avoid unnecessary Docker or network dependencies and end-to-end wiring is
  exercised once at the composition root.
- **When to skip it.** A module that only touches its own `domain` repository interface
  already has its port. Reserve Rule 10 for infrastructure that is not a shared domain
  repository: storage, a third-party service, or another module's narrow behavior.

For ports-and-adapters theory see the Hexagonal guide in `golang-design-patterns`; for
interface mechanics (small interfaces, structural satisfaction, compile-time checks) see
`golang-structs-interfaces`. Apply dependency inversion at infrastructure seams without
adding a `ports/` and `adapters/` directory scaffold unless it improves the repository's
actual dependency graph.

#### Plan & review checklist

Use this when planning a new module (enumerate the seams up front) and when reviewing a change (each item is a fail condition):

- [ ] Does business orchestration import an infra **concrete** or SDK instead of a local
      port? Repository adapters may use the established platform query abstraction.
- [ ] Is every external behavior used by business logic reachable through a focused local
      interface or an established shared `domain` contract?
- [ ] Are concrete clients constructed in `app/` and injected? Use `rg` for
      `storage.New`, `sql.Open`, or SDK constructors inside business logic.
- [ ] Can each unit be tested with a fake for every seam, without Docker or live network
      services?
- [ ] Are ports named for behavior and kept unexported unless a sibling package genuinely needs them?

## Verifying the layout

Run these after any structural change — a clean refactor is behavior-preserving.

Run the repository's automated boundary check, then its generation, build, test, and lint
commands:

```bash
make architecture-check
make check
```

Use these only as diagnostic spot-checks when the automated target reports a violation:

```bash
# informational: review each sibling-module import; established reviewed edges are allowed
rg 'internal/modules/[a-z]+"' internal/modules -g '*.go' -g '!**/*_test.go'
# hard fail: a module importing generated transport (should return nothing)
rg -l 'internal/api/' internal/modules -g '*.go'
# hard fail: platform reaching up into domain or modules (should return nothing)
rg -l 'internal/(modules|domain)' internal/platform -g '*.go'
```

## Migrating an existing flat `internal/`

A regroup is a large but low-risk diff — every domain import path changes, and the compiler catches every miss. Sequence it as **independently reviewable, revertible commits**:

1. **Fix mislabeled packages first** (e.g. a `repository/sqlbuilder` that is really a DB utility → `platform/database/sqlbuilder`; delete dead shells).
2. **`platform/` regroup** — move infrastructure packages only.
3. **`modules/` regroup** — move business domains only; fold single-consumer tooling (e.g. `seed`) into its consumer (`cli`).
4. Update the project's vendor-neutral `AGENTS.md` layout section and run generation,
   tests, and lint after each independently reviewable step.

Do this **before** more code piles on — it is the cheapest the move will ever be.

## Deciding against heavyweight architectures

Before adopting Clean/Hexagonal/"feature modules", check what the codebase already has:

- Repository **interfaces** in `domain` + injected concrete impls **is already dependency inversion** — the core of Hexagonal. You likely don't need ports/adapters ceremony on top.
- Package-by-feature is genuinely popular and good in Go — but the popular version keeps a central shared-types package (this skill's `domain`), it does not dissolve it.
- A folder diagram labeled "Hexagonal" that has a service holding a concrete `*Store` and a handler importing generated transport is **not** actually Hexagonal — it's package-by-feature with layering vocabulary. Take the good grouping instinct; skip the mislabel and the dependency spread.

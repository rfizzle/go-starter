# Module and Platform Boundaries

> Status: accepted · Date: 2026-07-23

## Context

As a starter grows, a flat `internal/` directory makes technical plumbing and
business capabilities look interchangeable. That encourages controllers to
accumulate business rules, infrastructure to reach into domain packages, and
new catch-all packages to become accidental dependency hubs.

This repository adopts the module/platform structure proven in Overwatch while
keeping Go's package model and explicit dependency injection. The grouping
directories communicate ownership; they do not introduce a framework.

## Decision

```text
internal/
  api/v1/                 generated ogen transport; never edit
  app/                    composition root and lifecycle
  auth/                   cross-cutting authentication edge
  cli/                    command support and development seeding
  controller/             generated transport ↔ domain mapping
  domain/                 shared business data and contracts
  modules/                business capabilities
    user/                  example PostgreSQL-backed user module
  platform/               technical infrastructure
    config/
    database/postgres/
    middleware/
    observability/
    ratelimiter/
    web/
  templateinit/           repository-initializer implementation
```

`internal/templateinit` remains top-level because it supports the standalone
template initialization tool; it is neither runtime infrastructure nor a
business capability.

### Import direction

```text
                         ┌──────────────┐
                         │     app      │
                         └──┬────┬────┬─┘
                            │    │    │
                 ┌──────────┘    │    └──────────┐
                 ▼               ▼               ▼
          controller/edge     modules/*       platform/*
                 │             │     │
                 ▼             ▼     ▼
            api/v1 + auth    domain  selected platform packages

domain: independent internal leaf
platform: independent of domain and modules
```

The rules are:

1. `domain` holds business data, shared repository interfaces, and
   business-facing sentinel errors. It imports no other application package.
2. `platform` holds infrastructure with no business meaning. It never imports
   `domain` or `modules`.
3. `modules` holds business capabilities. Modules may import `domain` and
   selected `platform` packages, but never generated transport.
4. Business orchestration declares small consumer-side ports. The composition
   root constructs concrete adapters and injects them.
5. A module's repository adapter may import the platform database boundary.
   Business logic above that adapter depends on a domain contract or a local
   consumer interface.
6. `controller` maps generated request/response types to domain calls. It must
   not become a second business layer or depend on a concrete module.
7. Generated API imports stay at the transport edge and composition root.
8. Packages named `utils`, `helpers`, `common`, or `shared` are prohibited.
   Code belongs to a named business capability or platform service.

The example user flow demonstrates the intended separation:

```text
controller.UserReader ──→ domain.User
                              ▲
                              │ implements domain.UserRepository
                       modules/user.Repository
                              │
                              ▼
               platform/database/postgres/db
```

The controller and CLI seeder consume `domain.User` and `domain.ErrNotFound`;
they do not import the concrete user module. `internal/app` and the seed command
are the only places that construct `user.Repository`.

## Placement guide

| Change | Location |
| --- | --- |
| Business model or shared repository contract | `internal/domain/` |
| Business repository, validation, or orchestration | `internal/modules/<capability>/` |
| PostgreSQL lifecycle, migrations, or generated SQL | `internal/platform/database/` |
| Config, logging, telemetry, middleware, rate limiting, HTTP serving | matching `internal/platform/<capability>/` package |
| OpenAPI request/response mapping | `internal/controller/` |
| Dependency construction or lifecycle ordering | `internal/app/` or a CLI composition root |
| Generated ogen code | `internal/api/v1/` via `make generate` |

Create a service layer inside a module only when real business orchestration
requires it. Do not create `ports/`, `adapters/`, factories, or interfaces for
single concrete implementations without a current substitution boundary.

## Enforcement

`make architecture-check` uses the Go package graph, including test imports, to
enforce:

- no platform dependency on domain or modules;
- no internal dependency from domain;
- no generated API dependency from modules;
- generated API imports only at approved edges;
- no Go package outside the approved top-level `internal/` groups; and
- no forbidden grab-bag package directory.

The target is part of `make check`. The repository-local `go-project-layout`
skill supplies the planning and review checklist, `backend-spec` governs
business module contracts, and `project-conventions` routes cross-cutting work
to this decision.

## Consequences and non-goals

- Package ownership and dependency direction are visible in paths and checked
  automatically.
- Structural changes require import updates, but the compiler and architecture
  check expose misses.
- The central domain package can grow; keep it flat and limited to genuinely
  shared business contracts.
- This decision does not mandate full Clean Architecture, a service per
  endpoint, or an interface per type.
- This change does not alter HTTP, persistence, authentication, seeding, or
  lifecycle behavior.

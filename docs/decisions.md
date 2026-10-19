# Decisions

What was decided, when and why. A record is never deleted; a superseded one
gains a status that names its successor. Ids are permanent and never reused.

## D1: Module and platform boundaries

**Status:** accepted · **Date:** 2026-07-23

**Context.** As a starter grows, a flat `internal/` directory makes technical
plumbing and business capabilities look interchangeable. That encourages
controllers to accumulate business rules, infrastructure to reach into domain
packages, and catch-all packages to become accidental dependency hubs. This
repository adopts the module/platform structure proven in Overwatch while
keeping Go's package model and explicit dependency injection. The grouping
directories communicate ownership; they do not introduce a framework.

**Decision.** `internal/` is split by ownership, and imports run one way. The
directory map is in `AGENTS.md`; the rules are these:

1. `domain` holds business data, persistence inputs and business-facing
   sentinel errors. It imports no other application package.
2. `platform` holds infrastructure with no business meaning. It never imports
   `domain` or `modules`.
3. `modules` holds business capabilities. A module may import `domain` and
   selected `platform` packages, never generated transport.
4. Business orchestration declares small consumer-side ports. The composition
   root builds the concrete adapters and injects them.
5. A module's repository adapter may import the platform database boundary.
   Business logic above that adapter depends on a domain contract or a local
   consumer interface.
6. `controller` maps generated request and response types to domain calls. It
   must not become a second business layer or depend on a concrete module; when
   it needs a module's behaviour it declares the port it consumes and the
   composition root adapts the module to it.
7. Generated API imports stay at the transport edge and the composition root.
8. Packages named `utils`, `helpers`, `common` or `shared` are prohibited.
   Code belongs to a named business capability or platform service.

`internal/templateinit` stays top-level because it supports the standalone
template initializer; it is neither runtime infrastructure nor a business
capability. `make init` deletes it with `tools/initialize`, so derived
projects do not carry it.

Create a service layer inside a module only when real business orchestration
requires it. Do not create `ports/` or `adapters/` packages, factories, or
interfaces for a single concrete implementation without a current
substitution boundary.

`make architecture-check` enforces the rules from the Go package graph,
including test imports, and is part of `make check`. The `go-project-layout`
skill is the checklist for placing a change.

**Consequences.**

- Package ownership and dependency direction are visible in paths and checked
  automatically.
- Structural changes require import updates, but the compiler and the
  architecture check expose misses.
- The central domain package can grow; keep it flat and limited to genuinely
  shared business contracts.
- The decision does not mandate full Clean Architecture, a service per
  endpoint or an interface per type.

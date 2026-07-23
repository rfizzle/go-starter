---
name: project-conventions
description: "Go Starter-specific module/platform architecture, generated transport, database, composition-root, and verification conventions. Use for cross-cutting changes, package moves, new business modules or infrastructure packages, or whenever repository ownership and dependency direction are unclear. Not for narrow in-package edits already governed by a more specific skill."
---

# Go Starter Project Conventions

Read the accepted
[module/platform boundary decision](../../../docs/architecture/module-platform-boundaries.md)
before making a cross-cutting change involving package placement, import direction,
generated transport, database ownership, controller responsibilities, or dependency
composition.

Apply that document as project policy. Then use:

- `go-project-layout` for placement, dependency seams, and structural review;
- `backend-spec` for new or behaviorally changed business modules;
- a focused `golang-*` skill for the implementation concern; and
- `make architecture-check` before the relevant tests and `make check`.

Do not load broad project guidance for a narrow change whose package and dependency
direction are already established.

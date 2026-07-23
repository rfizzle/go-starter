# Architecture Documentation

This directory records the structure and decisions that constrain Go Starter.

## Documents

- [Module and platform boundaries](module-platform-boundaries.md) — accepted
  package placement, import direction, composition, and automated enforcement.
- [Current architecture state](current-state.md) — the implemented runtime and
  package map.

## Updating architecture

Update the accepted boundary document when package ownership or permitted
dependency direction changes. Update current state when implemented
capabilities or wiring change. Significant changes should explain context,
decision, consequences, and rejected complexity.

Related guidance:

- [Database patterns](../development/database.md)
- [Interface patterns](../guidelines/interface-pattern.md)
- [Testing strategy](../guidelines/testing-strategy.md)
- Repository-local `go-project-layout` and `backend-spec` skills

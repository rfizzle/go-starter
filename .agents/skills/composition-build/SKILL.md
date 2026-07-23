---
name: composition-build
description: >
  Build a governed multi-component composition (a pattern that combines 2+ atomic
  components into a cohesive unit with its own data flow and interaction logic). Trigger
  after its child components exist and a composition spec has been approved. Produces the
  Svelte 5 composition, its Storybook story with realistic data, and integration tests.
---

# Composition Build (Svelte 5 + Tailwind 4)

A composition is not a primitive — it is a governed assembly of existing components that
together form a reusable UI pattern. Examples: a page header with actions, a filter panel,
a card with embedded status + metadata, a detail sidebar.

## Key differences from atomic components
- **No ARIA primitive backing** — compositions delegate a11y to their children.
- **Own state management** — may hold local state, derive values, call APIs.
- **Integration-focused testing** — tests verify data flows between children correctly,
  not atomic behavior (children already have their own tests).
- **Realistic Storybook stories** — show the composition with production-shaped data,
  multiple scenarios, not just variant grids.

## Output location
All artifacts live in one folder:

```
web/src/lib/compositions/<PascalName>/
├── <PascalName>.svelte
├── <PascalName>.stories.ts
├── <PascalName>.test.ts
├── <PascalName>.spec.md        (composition spec — simpler than component spec)
└── index.ts                    (barrel export)
```

## Procedure
1. Read the approved composition spec produced by `composition-spec`. Stop if it is
   missing or still `draft`. Identify the child components and confirm they exist.
2. Read `references/composition-conventions.md` for structural patterns.
3. Implement the approved props API:
   - Data props: the domain data the composition needs (typed from `$lib/api/types`)
   - Callback props: event handlers it emits upward (`on<action>`)
   - Slot props: snippets for caller-provided content areas
   - `class?`: for layout-level overrides (margin, grid placement)
4. Implement with Svelte 5 idioms:
   - `$props()` for the typed interface
   - `$state`/`$derived` for internal state
   - Import and compose child components
   - Style color/domain roles with semantic tokens and geometry with configured scales
   - Never re-style child component internals — use their prop APIs
5. Write the Storybook story (see `../component-build/references/storybook-conventions.md`):
   - Default story with realistic production data
   - Loading state story
   - Empty state story
   - Error state story (if applicable)
   - Interactive story demonstrating key user flows
6. Write integration tests:
   - Verify child components receive correct props from parent data
   - Verify callback props fire with correct payloads on user interaction
   - Verify derived state updates when source data changes
   - Verify loading/empty/error state rendering
7. Run tests. Composition is done only when green.

## Gate / Definition of Done (hard fail)
- Spec status is `approved` (the `check-specs` gate blocks the build otherwise).
- All required spec sections present: Purpose, Children, Props API, Internal State,
  Interaction Flows, Layout, Testing Scenarios.
- All child components exist and are imported (not inlined/reimplemented).
- Props API typed; all data flows through props, not global state or context hacks.
- Semantic color/domain tokens and configured layout/type/size scales; no arbitrary
  visual values.
- Child component internals not overridden (no reaching into their DOM).
- Storybook story shows realistic data for default, loading, empty, and error states.
- Integration tests verify data flow and event propagation between parent and children.
- All tests green.
- `node web/scripts/check-specs.mjs <composition-folder>` passes.

## Guardrails
- Never reimplement child component behavior — compose, don't clone.
- Never introduce raw visual values. Use semantic tokens for meaning and configured
  scales for ordinary geometry.
- Never use global state or Svelte context unless the feature brief explicitly requires it.
- Never test child component internals — their own test suite covers that.
- If you need a child component to do something it doesn't support, flag it as a
  component enhancement, don't work around it.

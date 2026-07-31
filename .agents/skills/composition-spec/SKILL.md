---
name: composition-spec
description: Create a governed Svelte 5 composition contract before implementation. Use when a reusable UI pattern combines two or more existing components and needs explicit children, props, state, interaction flows, layout, responsive behavior, accessibility delegation, and integration-test scenarios. Produces a draft spec for human approval and never writes implementation code.
---

# Composition Spec

Define the contract for a reusable assembly of existing components. Keep the spec at the
behavior and data-flow level; do not write Svelte implementation code.

## Output

Write the spec to:

```text
web/src/lib/compositions/<PascalName>/<PascalName>.spec.md
```

Use [the composition template](references/spec-template.md).

## Generic-first policy
Default to a **generic** composition: parameterized props and children, reusable across
more than one view or data shape. Before drafting a new spec, check
`web/src/lib/compositions/` for an existing composition that already fits, or that could
with a new prop/slot — extend it rather than spec a new one. Only spec a **specific
(typed)** one-off when the flow is high-value or extreme enough that generalizing it
would visibly harm that experience, and record the reason in Purpose.

## Procedure

1. Read the governing feature brief and inspect the existing component and composition
   inventory (`web/src/lib/components/`, `web/src/lib/compositions/`) for a fit or
   near-fit before proposing a new composition.
2. Confirm every required child exists or record it as a blocking dependency.
3. Fill every template section with concrete behavior and typed Svelte 5 props.
4. Describe accessibility owned by the composition: reading order, focus sequence,
   landmarks, announcements, and coordination between child widgets. Do not restate
   atomic child contracts.
5. Use semantic tokens for color/domain meaning and configured Tailwind scales for
   layout, typography, and sizing. Do not include raw visual values.
6. Write the spec with `Status: draft`.
7. Run `node .agents/skills/composition-spec/scripts/validate-spec.mjs <spec-path>`.
8. Present the validated spec for human approval, then stop. Do not implement and do not
   set `Status: approved` without explicit approval.

## Gate

- Purpose and reuse boundary are explicit; if specific rather than generic, Purpose
  states the high-value or extreme-UX reason.
- Existing composition inventory checked; no existing composition/prop already covers this.
- Children table names real components and data/event seams.
- Props API is typed and no broader than the composition needs.
- Internal state and every user interaction are documented.
- Layout, responsive behavior, and accessibility delegation are testable.
- Testing scenarios cover data flow, callbacks, loading, empty, and error states as
  applicable.
- Status remains `draft` until a human approves it.

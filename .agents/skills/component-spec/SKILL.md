---
name: component-spec
description: >
  Produce a complete component contract spec BEFORE any UI is built, for a Svelte 5 +
  Tailwind 4 + Bits UI project. Trigger whenever a new UI component, a variant of an
  existing one, or formalization of an existing component is requested. Resolves all
  design decisions to semantic design tokens and names the governing ARIA APG pattern.
  Never invents raw values and never writes implementation code.
---

# Component Spec

The spec is the contract that gates tests and code. No implementation may begin until a
spec passes its gate.

## Output location
Write the completed spec co-located with the component it describes:

```
web/src/lib/components/<PascalName>/<PascalName>.spec.md
```

Create the component directory if it does not exist. The spec, its tests, and its
implementation all live in this one folder so they review and version as a unit.
`<PascalName>` is the component name in PascalCase (e.g. `Button`, `DatePicker`).
(To use a central spec store instead, change this one path to `specs/<PascalName>.md`.)

## Procedure
1. Read `references/spec-template.md` and fill EVERY section. No section may be omitted.
2. Identify the governing native semantic or **ARIA APG pattern**. Open the official
   source routed from `../a11y-review/references/wcag22-aa-checklist.md`, record its URL,
   and summarize the complete required keyboard model. Mark optional keys as optional;
   do not copy the checklist's abbreviated routing note as the contract.
3. Identify the **Bits UI primitive** that should back the component (Dialog, Popover,
   Select, Combobox, Tabs, Accordion, Menu, Tooltip, Slider, Switch, Checkbox,
   RadioGroup, Toggle, …). If none exists, note the Melt UI builder fallback, or flag
   that a fully manual a11y contract is required.
4. Express the props API as a Svelte 5 `$props()` interface (typed; required vs optional
   with defaults; include `class?` for token-based overrides and `children?: Snippet`).
5. Reference semantic tokens for color and domain roles (e.g. `bg-surface`, `text-fg`,
   `ring-focus`) and configured Tailwind scales for layout, type, and size. If a semantic
   role is missing, STOP and use `design-tokens`; do not invent a value.
6. Write the spec to the Output location above with `Status: draft`.
7. Run the validator from the repository root and do not hand off until it exits 0:

   ```bash
   node .agents/skills/component-spec/scripts/validate-spec.mjs <path-to-spec.md>
   ```
8. Present the draft for human approval. Do not change `Status` to `approved` on the
   agent's own authority.

## Status field (canonical)
The status line vocabulary is exactly `draft | approved` — no `reviewed`, `complete`,
`wip`, or blank. A new spec is always written as `draft`. Only a human reviewer flips it
to `approved`, and only then may the component be built. The `check-specs` gate
(`web/scripts/check-specs.mjs`, wired into `make web-build` and `make ci`) fails the
build for any built `.svelte` unit whose co-located spec is missing, non-approved, or
structurally invalid — so an unapproved spec can never ship silently.

## Gate (hard fail)
- Spec written to `web/src/lib/components/<PascalName>/<PascalName>.spec.md`.
- Every template section present and non-placeholder.
- A named native semantic or APG pattern, governing source, and full keyboard model.
- A named backing primitive (or explicit justified fallback).
- Props API typed; all interactive states enumerated
  (default, hover, focus-visible, active, disabled, loading, error, selected as applicable).
- Zero raw visual values; colors/domain roles use semantic tokens and ordinary geometry
  uses configured Tailwind scales.
- Status is `draft` until a human approves; the build gate requires `approved`.

## Guardrails
- Never write `.svelte` implementation here — spec only.
- Never invent tokens, colors, spacing, or keyboard behavior; resolve to the
  token layer and the APG pattern.

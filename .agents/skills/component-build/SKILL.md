---
name: component-build
description: >
  Implement and test a Svelte 5 + Tailwind 4 component from an approved component spec.
  Use after spec accessibility review and human approval. Produces specification-derived
  tests, the `.svelte` implementation, and Storybook stories using native HTML or Bits UI
  as appropriate, semantic color tokens, and configured Tailwind scales.
---

# Component Build (Svelte 5 + Tailwind 4 + Bits UI)

Implementation is written to a fixed target: the approved spec. Do not begin while the
spec is missing, draft, or materially inconsistent with the requested behavior.

## Procedure
1. Read the approved spec at `web/src/lib/components/<PascalName>/<PascalName>.spec.md`,
   `references/svelte5-tailwind4-conventions.md`, and `references/storybook-conventions.md`.
   For a Bits UI component, also read `references/bits-ui-testing.md` and verify the
   installed primitive API rather than assuming the reference matches every version.
   Write all outputs into that same folder:
   `<PascalName>.svelte`, `<PascalName>.stories.ts`, `<PascalName>.test.ts`.
2. Write or update tests from the approved contract before implementation. Cover props,
   states, keyboard interaction, callbacks, and accessible name/role/state. Reuse existing
   tests when they already express the contract.
3. Prefer native elements for native behavior (`button`, `input`, `select`, `details`).
   For **complex interactive patterns** (dialogs, comboboxes, tabs, accordions, sliders,
   popovers, date pickers), compose the matching Bits UI primitive — do not re-implement
   its behavior. Use Melt UI only where Bits UI has no primitive.
   Do not hand-roll focus traps, composite-widget navigation, or ARIA state machines.
4. Use Svelte 5 idioms: `$props()` with the spec's typed interface, `$state`/`$derived`
   for local state, `$bindable()` for two-way props, snippets (`{@render}`) for slots,
   event handlers as props (`onclick`, not `on:click`).
5. Style colors and domain roles with semantic utilities (`bg-surface`, `text-fg`,
   `ring-focus`). Use configured Tailwind scales for layout, typography, and sizing.
   Merge the incoming `class` prop last via `cn()` so overrides win safely.
6. Implement focus-visible (`focus-visible:ring-2 focus-visible:ring-focus`), reduced
   motion (`motion-reduce:*`), disabled/loading states per spec.
7. Do not suppress Svelte accessibility compiler warnings.
8. Run focused tests and repository checks. Run axe when Storybook and its browser are
   available; otherwise report it as unverified.

## Gate / Definition of Done (hard fail)
- Uses the declared native element or primitive.
- Semantic colors/domain roles and configured scale utilities; no arbitrary values or
  built-in palette colors (verified by source review; the contrast validator does not scan
  component class strings).
- All spec states implemented and themed; contrast holds in every theme.
- Interaction and accessibility tests green; run visual-regression tests only when the
  repository has a configured snapshot/visual service. `axe` has zero WCAG A/AA violations.
- Keyboard-complete; focus visible + returns correctly; reduced-motion honored.
- No suppressed Svelte a11y warnings; required Storybook stories exist.

## Guardrails
- Never introduce raw visual values. Request a semantic token for missing meaning; use
  existing configured scales for ordinary geometry.
- Never hand-roll complex focus management (traps, roving tabindex) when a primitive exists.
  Prefer native elements for simple behavior rather than recreating an ARIA widget.
- Never widen the props API beyond the approved spec without a spec revision.

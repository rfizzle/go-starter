---
name: feature-decompose
description: >
  Decompose a UX requirement into a governed build plan — inventorying tokens, components,
  compositions, data needs, and interaction flows. Trigger when a new page, view, or
  multi-component feature is requested. Produces a feature brief that gates all downstream
  build work. Never writes implementation code.
---

# Feature Decompose

The architect step. Given a UX requirement (a page, a workflow, a panel), produce a
structured inventory of everything that must exist before the feature can ship — and what
order to build it in.

## Output location
Write the completed brief to:

```
web/src/features/<feature-slug>/<feature-slug>.brief.md
```

Create the directory if it does not exist. The brief, its compositions, and its view
assembly all live in this folder.

## Generic-first policy
Default every new unit to **generic**: parameterized, reusable across contexts, not bound
to one screen's exact data shape. Reach for a **specific (typed)** one-off only when the
interaction is high-value or otherwise extreme enough (a primary conversion or safety
flow, a signature brand moment) that generalizing it would visibly compromise that user
experience. A specific unit still needs a one-line reason it can't be generic — not a
blanket default.

Before proposing anything new, check whether an existing generic component or composition
already covers the need, or could with a new variant/prop. Extension beats a new unit;
a new generic unit beats a new specific one.

## Procedure
1. Read the UX requirement. If it references a spec doc, read that too.
2. Read the existing component and composition inventory at `web/src/lib/components/`
   and `web/src/lib/compositions/`. Identify what already exists and can be reused as-is,
   and what could satisfy this requirement with a new variant or prop instead of a new unit.
3. Fill the template from `references/brief-template.md`. Every section is required.
4. For each element in the inventory, classify as:
   - **Exists** — already built; reference its path
   - **Needs enhancement** — exists but needs a variant or prop addition; prefer this over
     a new unit whenever the existing component/composition's purpose still fits
   - **New generic component** — atomic, parameterized for reuse; will go through
     `component-spec`, `a11y-review`, then `component-build`. The default for new units.
   - **New specific component** — atomic, bound to one exact use case; same pipeline as
     above, but the brief must state why generalizing it would harm the target UX
   - **New generic composition** — multi-component pattern usable across views; will go
     through `composition-spec`, then `composition-build`
   - **New specific composition** — multi-component pattern for one flow only; same
     pipeline, with the same one-line justification requirement as a specific component
5. Define the **interaction map**: what user actions trigger what state changes, API calls,
   or navigation. Focus on data flow, not visual layout.
6. Define **testing scenarios**: golden path, empty states, error states, loading states,
   edge cases (pagination boundaries, long strings, RTL, etc).
7. Propose a **build order** that respects dependencies (tokens first, then components,
   then compositions, then view assembly).

## Gate (hard fail)
- Brief written to the output location.
- Every section of the template filled with specifics, not placeholders.
- All UI elements classified (exists / needs enhancement / new generic component /
  new specific component / new generic composition / new specific composition).
- Every "specific" classification carries a one-line reason generic wasn't viable.
- Interaction map covers every user-initiated action on the page.
- Build order respects the dependency graph (no composition built before its children).
- Testing scenarios cover golden path + at least 3 edge cases.

## Guardrails
- Never write implementation code — brief only.
- Never skip the existing-inventory check. Reuse and extension are the default.
- Never propose a new component when a composition of existing ones would suffice.
- Never propose a specific (typed) unit without a stated high-value or extreme-UX reason;
  default to generic.
- Never propose building compositions before their child components exist or are planned.
- If a needed token doesn't exist, note it as a dependency — don't invent values.

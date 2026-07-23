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

## Procedure
1. Read the UX requirement. If it references a spec doc, read that too.
2. Read the existing component inventory at `web/src/lib/components/` and identify what
   already exists and can be reused as-is.
3. Fill the template from `references/brief-template.md`. Every section is required.
4. For each element in the inventory, classify as:
   - **Exists** — already built; reference its path
   - **New component** — atomic; will go through `component-spec`, `a11y-review`, then
     `component-build`
   - **New composition** — multi-component pattern; will go through `composition-spec`,
     then `composition-build`
   - **Needs enhancement** — exists but needs a variant or prop addition
5. Define the **interaction map**: what user actions trigger what state changes, API calls,
   or navigation. Focus on data flow, not visual layout.
6. Define **testing scenarios**: golden path, empty states, error states, loading states,
   edge cases (pagination boundaries, long strings, RTL, etc).
7. Propose a **build order** that respects dependencies (tokens first, then components,
   then compositions, then view assembly).

## Gate (hard fail)
- Brief written to the output location.
- Every section of the template filled with specifics, not placeholders.
- All UI elements classified (exists / new component / new composition / needs enhancement).
- Interaction map covers every user-initiated action on the page.
- Build order respects the dependency graph (no composition built before its children).
- Testing scenarios cover golden path + at least 3 edge cases.

## Guardrails
- Never write implementation code — brief only.
- Never skip the existing-inventory check. Reuse is the default.
- Never propose a new component when a composition of existing ones would suffice.
- Never propose building compositions before their child components exist or are planned.
- If a needed token doesn't exist, note it as a dependency — don't invent values.

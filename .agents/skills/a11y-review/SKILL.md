---
name: a11y-review
description: >
  Audit a component/composition spec or built Svelte 5 UI against WCAG 2.2 AA and the ARIA
  Authoring Practices Guide. Trigger after a spec is drafted (spec review) and after a
  component is built (code review), or whenever accessibility, keyboard support, focus,
  contrast, or screen-reader behavior is in question. Produces findings + a redline and
  a hard pass/fail gate.
---

# Accessibility Review

Runs at two points: against the **spec** (cheap, catches design-level gaps) and against
the **built component** (catches implementation gaps). Both must pass.

## Procedure
1. Load `references/wcag22-aa-checklist.md`.
2. **Spec review:**
   - For an atomic component, confirm the spec names the native semantic or APG pattern,
     links its governing source, summarizes required keyboard behavior, defines focus
     management, and specifies screen-reader name/role/state/value.
   - For a composition, confirm child widget responsibilities are delegated explicitly
     and review DOM reading order, focus sequence, landmarks, announcements, and keyboard
     transitions between children. Do not require one APG pattern for the whole assembly.
3. **Code review (Svelte):**
   - Confirm the UI uses its declared native element or established primitive. Hand-rolled
     interactive widgets require explicit justification and the strictest scrutiny.
   - Confirm NO Svelte a11y compiler warnings are suppressed (`<!-- svelte-ignore a11y-* -->`
     is a finding unless explicitly justified and approved).
   - Run `node web/scripts/axe-scan.mjs` against the component's stories when Storybook
     and the required browser are available. Record an unavailable runtime as unverified,
     never as a pass.
   - Verify focus-visible styling uses `ring-focus`; verify reduced-motion handling.
   - Verify target size ≥ 24×24 CSS px unless a documented WCAG 2.5.8 exception applies,
     and verify applicable foreground/background contrast in every theme.
4. Produce findings as a redline: each item tagged `blocker | major | minor` with the
   WCAG SC or APG rule it violates and the fix.

## Gate (hard fail)
- Zero unresolved `blocker`/`major` findings.
- `axe` reports zero WCAG A/AA violations across all audited stories/states; impact is
  reported for prioritization, not used to waive conformance.
- Keyboard model fully operable with no mouse; focus always visible; focus returns
  correctly after overlays close.
- All applicable contrast pairings meet 2.2 AA in every supported theme.

## Guardrails
- Do not "pass" anything by reframing it as out of scope; log it as a finding.
- Automated axe passing is necessary but NOT sufficient — manual keyboard + SR review
  items in the checklist still apply.
- Keep AAA findings advisory unless the product contract explicitly makes them mandatory.

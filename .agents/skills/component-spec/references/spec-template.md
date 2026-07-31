# `<Component>` Spec

> Status: draft · Spec version: 2 · Owner: · APG/native pattern: · Source: · Backing primitive:
>
> Status vocabulary is exactly `draft | approved` — no other values. A unit may only be
> built once its spec reads `Status: approved`. The `check-specs` gate enforces this.

## 1. Purpose
- What it is, when to use, when NOT to use (and what to use instead).
- **Generic or specific:** state which, and if specific, the one-line high-value or
  extreme-UX reason it isn't generalized.

## 2. Anatomy
- Named parts (e.g. root, trigger, content, label, icon, indicator).

## 3. Variants & sizes
- Variants: e.g. `primary | secondary | ghost | danger`
- Sizes: e.g. `sm | md | lg`  (state min target size ≥ 24×24 CSS px)

## 4. States
For each: default, hover, focus-visible, active, disabled, loading, error, selected
(as applicable) — describe visual treatment **by semantic token**, not raw value.

## 5. Props API (Svelte 5 `$props`)
```ts
type Props = {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'; // default 'primary'
  size?: 'sm' | 'md' | 'lg';                               // default 'md'
  disabled?: boolean;                                      // default false
  // ...
  class?: string;                                          // token-based override hook
  children?: import('svelte').Snippet;
};
```

## 6. Accessibility contract
- **Governing native/APG pattern:** `<pattern-name>`
- **Source:** `<official URL or native HTML specification>`
- **Semantic element / role:** e.g. native `<button>` / `role="dialog"`
- **Keyboard model:** (summarize the governing source; distinguish required and optional keys)
  | Key | Behavior |
  |-----|----------|
- **Focus management:** initial focus · focus trap? · return focus on close
- **Screen-reader expectations:** name / role / state / value announced
- **Target size:** ≥ 24×24 CSS px hit area
- **Reduced motion:** behavior under `prefers-reduced-motion`

## 7. Content & i18n
- Label length / truncation rules · RTL behavior · pluralization / number/date formatting.

## 8. Responsive behavior
- Behavior per breakpoint (reference `--breakpoint-*` tokens).

## 9. Tokens used and configured scales
- Enumerate semantic color/domain tokens and configured layout/type/size utilities. No
  arbitrary values.

## 10. Open questions
- Decisions deferred to review.

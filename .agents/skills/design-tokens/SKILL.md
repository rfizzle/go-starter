---
name: design-tokens
description: >
  Generate and validate the design-token layer for a Svelte 5 + Tailwind 4 project.
  Trigger whenever tokens, colors, themes, spacing/motion scales, or a new theme are
  requested, or when another skill reports a missing token. Produces semantic tokens inside
  a Tailwind 4 `@theme` block with a dark theme override. Components use semantic tokens
  for color and domain meaning and configured Tailwind scales for ordinary geometry.
---

# Design Tokens (Tailwind 4)

The token layer is the single source of truth for semantic color, typography, radius,
shadow, motion, and domain roles. Components never contain raw color values or use
Tailwind's built-in color palette.

## Architecture (flat semantic layer)

Tokens are declared directly in `@theme` with inlined OKLCH values. No intermediate
primitive color ramp — the semantic name is the abstraction. Keep the vocabulary lean,
but let the application own the roles it actually needs.

```css
@theme {
  --color-surface: oklch(100% 0 0);      /* page background */
  --color-fg: oklch(20% 0.016 260);      /* primary text */
  --color-brand: oklch(55% 0.13 168);    /* primary action */
}
```

Components use `bg-surface`, `text-fg`, and `bg-brand`. The `.dark` override redefines
the same custom properties, so component code does not change.

## Token vocabulary

The token set covers these semantic roles:

| Category | Tokens | Purpose |
|----------|--------|---------|
| Surface | `surface`, `surface-raised`, `surface-sunken`, `surface-overlay` | Background layers |
| Foreground | `fg`, `fg-muted`, `fg-subtle`, `fg-on-brand`, `fg-on-fill` | Text and icons |
| Border | `border`, `border-strong` | Separators and outlines |
| Brand | `brand`, `brand-emphasis`, `brand-muted`, `brand-subtle` | Primary actions and accents |
| Focus | `focus` | Focus ring color |
| Intents | `danger`, `warning`, `success`, `info` + `-muted` and `-fg` variants | Semantic feedback |
| Domain | `risk-*`, `avatar-tint-*`, chart roles | App-specific visualization roles |
| Non-color | font/type, tracking, leading, spacing base, `radius-*`, `shadow-*`, motion, z-index | Shared scales and named decisions |

Use Tailwind v4's documented namespaces. For example, named `duration-fast` utilities
come from `--transition-duration-fast`, not `--duration-fast`.

## When to add a new token

Add a token when:
- A component needs a semantic color role that doesn't exist (e.g., a new intent)
- You need a new surface elevation or foreground weight
- A domain concept needs its own color vocabulary (like risk/severity)

Do not add a semantic token for:
- A one-off color that's really just an existing token with opacity (use `bg-brand/10`)
- A one-off spacing or size alias — use the configured scale (`p-4`, `gap-3`, `h-10`)
- A component-specific name that merely duplicates an existing semantic role

## Procedure
1. Identify the semantic need. Name it by role, not appearance.
2. Pick an OKLCH value and calculate contrast; lightness heuristics are not proof.
3. Add the token to `@theme` in `web/src/app.css`.
4. Add its dark override in the `.dark` block when the semantic role must change. If the
   exact value is intentionally theme-invariant, document that decision beside the token.
5. If the token is an intent (danger, warning, etc.), provide: base, `-muted`, and `-fg`.
6. Add every intended foreground/background pairing to
   `scripts/contrast.config.json`.
7. From the repository root, run:

   ```bash
   node web/scripts/validate-tokens.mjs
   ```

   Treat a failing pairing as a design issue. Do not weaken the configured level merely
   to make the command pass.

## Contrast rules (WCAG 2.2 AA)
- Body text on its background: **4.5:1 minimum**
- Large text (≥18px or ≥14px bold) and UI components: **3:1 minimum**
- `fg-on-brand` on `brand`: must pass 4.5:1 in both themes

## File location
All tokens live in one file: `web/src/app.css`. The structure is:
1. `@theme { ... }` — light theme (default) semantic tokens
2. `.dark { ... }` — dark theme overrides of the same custom properties
3. `@layer base { ... }` — base element styles

## Gate (hard fail)
- Every pairing declared in the explicit contrast contract meets WCAG 2.2 AA in both
  themes, and review confirms the contract lists every pairing introduced by the change.
- The validator discovers both the base `@theme` declarations and `.dark` overrides.
- No token resolves to a value that duplicates an existing token's role.
- Token name describes its role, not its appearance (no `color-mint-500`).
- A `.dark` theme exists. Every color token either has an override or is explicitly
  documented as theme-invariant; inherited values are still contrast-tested in dark mode.

## Guardrails
- Never emit tokens that reference Tailwind's built-in palette (`slate-900`, `blue-500`).
- Never let components use raw values — if they need something, add a token here.
- Palette/arbitrary-utility bans are a source-review gate; the contrast script validates
  token values and declared pairings, not component class strings.
- Do not add a token when a configured layout/type/size scale utility already expresses
  the decision.
- Keep the total token count lean. Resist creating tokens for slight variations; use
  opacity modifiers (`bg-brand/10`) or existing tokens at different weights first.
- OKLCH hue angles: keep consistent per intent (brand≈168, danger≈25, warning≈80,
  success≈145, info≈250). Shift lightness and chroma, not hue.

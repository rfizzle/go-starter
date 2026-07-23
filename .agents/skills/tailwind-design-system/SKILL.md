---
name: tailwind-design-system
description: Build and review Svelte 5 component styling with Tailwind CSS v4, semantic design tokens, responsive layouts, themes, state variants, and accessible motion. Use when creating or standardizing UI patterns, reviewing Tailwind classes, or evolving the component styling system. Use design-tokens for changes to the token layer itself.
---

# Tailwind Design System

Apply Tailwind CSS v4 consistently in the Svelte frontend. Treat the current `@theme`
block and existing components as the source of truth; do not introduce a parallel React,
Radix, CVA, or `tailwind.config.*` architecture.

## Styling policy

- Use semantic utilities for color and domain meaning: `bg-surface`, `text-fg`,
  `border-border`, `ring-focus`, `text-fg-on-brand`, and intent/risk tokens.
- Use the configured Tailwind scales for layout, spacing, size, typography, radius,
  shadows, and breakpoints: `gap-4`, `h-10`, `text-sm`, `rounded-md`, `md:grid-cols-2`.
- Do not use palette utilities (`blue-500`, `slate-900`), arbitrary values
  (`p-[13px]`, `bg-[#123456]`), or inline visual styles.
- If a semantic color role is missing, stop and use `design-tokens`; do not disguise a
  missing token with an arbitrary value.
- Merge a caller's `class` value last with the project's `cn()` helper.

## Workflow

1. Read `web/src/app.css` and the closest existing component or composition.
2. Identify which decisions are semantic colors and which are scale-based layout or
   typography choices.
3. Implement mobile-first structure, then add only the breakpoints the content needs.
4. Cover hover, focus-visible, active, selected, disabled, loading, empty, and error
   states that apply to the contract.
5. Pair transitions and animations with a `motion-reduce:` alternative.
6. Verify light and dark themes, keyboard focus, reflow at 320 CSS px, and long content.
7. Run the repository's frontend checks after any required generation preflight.

## Svelte class composition

Prefer typed lookup objects or simple conditional arrays. Keep component variants close
to the component unless the same governed variant contract is reused by multiple units.

```svelte
<script lang="ts">
  import { cn } from '$lib/utils/cn';
  import type { Snippet } from 'svelte';

  type Tone = 'neutral' | 'danger';
  let { tone = 'neutral', class: className = '', children }: {
    tone?: Tone;
    class?: string;
    children?: Snippet;
  } = $props();

  const tones: Record<Tone, string> = {
    neutral: 'bg-surface-raised text-fg border-border',
    danger: 'bg-danger text-danger-fg border-danger',
  };
</script>

<div class={cn('rounded-md border p-4', tones[tone], className)}>
  {@render children?.()}
</div>
```

Follow `svelte5-best-practices` for complete component API and snippet guidance.

## Responsive and state rules

- Let content determine breakpoints; do not create device-specific component forks.
- Prefer container queries for reusable units whose layout depends on their container.
- Keep DOM order aligned with reading and focus order; visual reordering must not make
  keyboard navigation confusing.
- Preserve visible focus in every state. Do not remove outlines without an equivalent
  `focus-visible` indicator.
- Use opacity modifiers only when contrast remains valid in both themes.

## Gate

- No raw color or arbitrary visual values.
- No framework examples or dependencies that are absent from the Svelte project.
- Semantic color roles come from the token layer.
- Layout and typography use configured scales consistently.
- Light/dark, reduced-motion, responsive, and focus states are verified.

Read [advanced patterns](references/advanced-patterns.md) only for container queries,
class-based dark mode, motion, and reusable Svelte class patterns.

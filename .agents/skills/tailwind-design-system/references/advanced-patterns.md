# Advanced Tailwind v4 Patterns for Svelte

## Contents

- [Theme boundaries](#theme-boundaries)
- [Container queries](#container-queries)
- [Dark mode](#dark-mode)
- [Motion](#motion)
- [Reusable class sets](#reusable-class-sets)
- [Migration checks](#migration-checks)

## Theme boundaries

Declare semantic colors and named design decisions in `@theme`. Components consume the
generated utilities and do not repeat raw CSS values.

Use configured scale utilities for ordinary geometry:

```svelte
<section class="grid gap-4 rounded-md border border-border bg-surface p-4 text-fg">
  <!-- content -->
</section>
```

Add a semantic token when meaning must survive a theme change. Do not add a semantic
token merely to rename `gap-4` or `text-sm`.

## Container queries

Use container queries when a reusable component can appear in sidebars and full-width
content areas:

```svelte
<article class="@container rounded-md border border-border bg-surface p-4">
  <div class="grid gap-3 @md:grid-cols-2">
    <!-- content -->
  </div>
</article>
```

Use viewport breakpoints for page-level layout and container queries for reusable units.

## Dark mode

The project uses a class-based variant and overrides the same semantic custom properties:

```css
@custom-variant dark (&:where(.dark, .dark *));

.dark {
  --color-surface: oklch(...);
  --color-fg: oklch(...);
}
```

Most components should not need `dark:` color classes because semantic tokens already
change value. Use `dark:` only for behavior that is genuinely theme-specific.

## Motion

Use named animation utilities from the theme and always supply a reduced-motion path:

```svelte
<div class="animate-slide-up motion-reduce:animate-none">
  <!-- content -->
</div>
```

Avoid continuous decorative animation. Never make understanding or operating a control
depend on animation.

## Reusable class sets

For a small finite variant API, use a typed lookup and `cn()`:

```ts
type Tone = 'neutral' | 'success' | 'danger';

const tones: Record<Tone, string> = {
  neutral: 'bg-surface-raised text-fg',
  success: 'bg-success-muted text-fg',
  danger: 'bg-danger-muted text-fg',
};
```

Introduce a variant library only when the repository already depends on one or repeated
contracts justify the dependency. Do not copy React CVA patterns into Svelte by default.

## Migration checks

When migrating Tailwind v3 code:

1. Replace `@tailwind` directives with `@import "tailwindcss"`.
2. Move supported theme configuration into CSS `@theme` declarations.
3. Replace palette utilities with the project's semantic colors.
4. Confirm the class-based dark variant matches the existing selector.
5. Replace animation plugins with named theme animations where practical.
6. Run the token validator and frontend checks.

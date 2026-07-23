# Svelte 5 + Tailwind 4 + Bits UI — Conventions

> Verify exact Bits UI component APIs and any Tailwind 4 var-shorthand syntax against
> current docs; both ecosystems iterate. Patterns below reflect the intended idioms.

## Class composition helper (`$lib/utils/cn.ts`)
```ts
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
export const cn = (...inputs: ClassValue[]) => twMerge(clsx(inputs));
```
`twMerge` resolves Tailwind conflicts so a caller's `class` prop reliably overrides
internal classes when merged last.

## Worked example — Button (native element, no primitive needed)
```svelte
<script lang="ts">
  import { cn } from '$lib/utils/cn';
  import type { Snippet } from 'svelte';
  import type { HTMLButtonAttributes } from 'svelte/elements';

  type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
  type Size = 'sm' | 'md' | 'lg';

  type Props = Omit<HTMLButtonAttributes, 'children'> & {
    variant?: Variant;
    size?: Size;
    loading?: boolean;
    children?: Snippet;
  };

  let {
    variant = 'primary',
    size = 'md',
    type = 'button',
    disabled = false,
    loading = false,
    class: className = '',
    children,
    ...rest
  }: Props = $props();

  // semantic colors + configured Tailwind geometry/type scales; no raw values
  const base =
    'inline-flex items-center justify-center gap-2 rounded-md font-medium ' +
    'transition-colors ease-out motion-reduce:transition-none ' +
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-2 focus-visible:ring-offset-surface ' +
    'disabled:opacity-50 disabled:pointer-events-none';

  const variants: Record<Variant, string> = {
    primary:   'bg-brand text-fg-on-brand hover:bg-brand-emphasis',
    secondary: 'bg-surface-raised text-fg border border-border hover:bg-surface-sunken',
    ghost:     'bg-transparent text-fg hover:bg-surface-sunken',
    danger:    'bg-danger text-danger-fg hover:bg-danger/90',
  };
  const sizes: Record<Size, string> = {
    sm: 'h-8 px-3 text-sm',   // 32px ≥ 24px target
    md: 'h-10 px-4 text-sm',
    lg: 'h-12 px-6 text-base',
  };
</script>

<button
  {type}
  disabled={disabled || loading}
  aria-busy={loading}
  class={cn(base, variants[variant], sizes[size], className)}
  {...rest}
>
  {#if loading}
    <span
      class="size-4 animate-spin rounded-full border-2 border-current border-t-transparent motion-reduce:animate-none"
      aria-hidden="true"
    ></span>
  {/if}
  {@render children?.()}
</button>
```
Notes: focus ring uses the `ring-focus` token; reduced motion via `motion-reduce:*`;
`aria-busy` exposes loading state; `class` merged last so callers can override.

## Worked example — primitive-backed (Bits UI Dialog)
```svelte
<script lang="ts">
  import { Dialog } from 'bits-ui';
  import { cn } from '$lib/utils/cn';
  import type { Snippet } from 'svelte';
  let { title, description, children }:
    { title: string; description?: string; children: Snippet } = $props();
</script>

<Dialog.Root>
  <Dialog.Trigger>
    {#snippet child({ props })}
      <button {...props} class="rounded-md bg-brand px-4 py-2 text-fg-on-brand">
        Open dialog
      </button>
    {/snippet}
  </Dialog.Trigger>
  <Dialog.Portal>
    <Dialog.Overlay class="fixed inset-0 bg-fg/40 animate-fade-in motion-reduce:animate-none" />
    <Dialog.Content
      class={cn(
        'fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2',
        'rounded-md border border-border bg-surface text-fg p-6 shadow-md',
        'focus-visible:outline-none'
      )}
    >
      <Dialog.Title class="text-fg font-semibold">{title}</Dialog.Title>
      {#if description}
        <Dialog.Description class="text-fg-muted text-sm">{description}</Dialog.Description>
      {/if}
      {@render children()}
      <Dialog.Close class="...">Close</Dialog.Close>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
```
Bits UI supplies the focus trap, focus return, `Esc` handling, `aria-modal`, and
title/description wiring — the build skill only styles with tokens and composes content.

## Forbidden patterns (token validation / review will reject)
- Raw values: `bg-[#1e40af]`, `p-[13px]`, `text-[oklch(...)]`, inline `style=`.
- Built-in palette colors: `bg-blue-500`, `text-neutral-900` (use semantic colors).
- Arbitrary geometry when a configured scale exists: `p-[13px]`, `h-[41px]`.
- Hand-rolled focus traps / roving tabindex when a Bits UI primitive covers the pattern.
  (Simple keyboard handling — arrows over a short option list, Escape to close — is fine without a primitive.)
- Suppressed a11y warnings: `<!-- svelte-ignore a11y-* -->` without approved justification.
- Legacy Svelte 4 idioms: `on:click`, `export let`, `$$props`, `<slot>` (use runes/snippets).

## Motion tokens
Reference motion tokens (`--ease-out`, named durations) via Tailwind's theme utilities
where generated, otherwise via `var()` in an arbitrary property. Always pair animation
with a `motion-reduce:` escape.

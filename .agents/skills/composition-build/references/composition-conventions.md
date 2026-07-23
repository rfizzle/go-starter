# Composition Conventions (Svelte 5 + Tailwind 4)

## Structural patterns

### Data-down, events-up
Compositions receive domain data as props and emit user actions as callback props.
They never fetch their own data (that's the view's job) unless they encapsulate a
self-contained interaction (e.g., an inline edit that PATCHes and updates local state).

```svelte
<script lang="ts">
  import type { Session } from '$lib/api/types';
  import StatusIndicator from '$lib/components/display/StatusIndicator.svelte';
  import UserChip from '$lib/components/display/UserChip.svelte';
  import TimeAgo from '$lib/components/display/TimeAgo.svelte';

  let {
    session,
    onaction,
    class: className = '',
  }: {
    session: Session;
    onaction?: (action: string) => void;
    class?: string;
  } = $props();
</script>
```

### Layout responsibility
A composition owns the *arrangement* of its children (flex, grid, gap, padding between
sections) but never the *internal styling* of children. Use configured spacing scales and
semantic border/color tokens between children.

```svelte
<header class="flex items-center justify-between gap-4 border-b border-border pb-4">
  <div class="flex items-center gap-3">
    <StatusIndicator status={session.status} />
    <UserChip email={session.user.email} />
  </div>
  <TimeAgo timestamp={session.started_at} />
</header>
```

### Snippet slots for extensibility
When a composition has regions that vary by context, use named snippets:

```svelte
<script lang="ts">
  import type { Snippet } from 'svelte';
  let { header, actions, children }: {
    header?: Snippet;
    actions?: Snippet;
    children: Snippet;
  } = $props();
</script>

<div class="flex flex-col gap-4">
  {#if header}{@render header()}{/if}
  {@render children()}
  {#if actions}
    <footer class="flex justify-end gap-2">{@render actions()}</footer>
  {/if}
</div>
```

### Internal state
Compositions can hold derived or local state that doesn't need to leak upward:

```svelte
<script lang="ts">
  let { items }: { items: Item[] } = $props();

  // Internal — parent doesn't need to know about expand/collapse
  let expandedId: string | null = $state(null);

  function toggle(id: string) {
    expandedId = expandedId === id ? null : id;
  }

  let sortedItems = $derived(items.toSorted((a, b) => b.score - a.score));
</script>
```

### Loading / empty / error states
Every composition that renders data must handle all three:

```svelte
{#if loading}
  <div class="animate-pulse motion-reduce:animate-none space-y-3">
    <div class="h-4 w-1/3 rounded bg-surface-sunken"></div>
    <div class="h-4 w-2/3 rounded bg-surface-sunken"></div>
  </div>
{:else if error}
  <div class="rounded-md border border-danger/20 bg-danger/5 p-4 text-danger">
    <p class="text-sm">{error}</p>
  </div>
{:else if items.length === 0}
  <EmptyState title="No items" description="Nothing to show." />
{:else}
  <!-- main content -->
{/if}
```

### File naming in the compositions directory
```
web/src/lib/compositions/
├── SessionHeader/
│   ├── SessionHeader.svelte
│   ├── SessionHeader.stories.ts
│   ├── SessionHeader.test.ts
│   ├── SessionHeader.spec.md
│   └── index.ts
├── StreamTimeline/
│   ├── ...
└── ReviewPanel/
    ├── ...
```

### Barrel exports
Each composition folder has an `index.ts`:
```ts
export { default as SessionHeader } from './SessionHeader.svelte';
```

### When to extract a composition vs. keep inline
Extract into `compositions/` when:
- Used in 2+ views
- Has its own internal state logic (expand/collapse, tab switching, filtering)
- Has 3+ child components arranged together
- Warrants its own Storybook entry for isolated development

Keep inline in the view when:
- One-off layout wiring with no internal logic
- Just arranging 1-2 components with some spacing
- Tightly coupled to the view's specific data shape with no reuse potential

### Testing compositions
Integration tests focus on the seams:

```ts
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import SessionHeader from './SessionHeader.svelte';

const mockSession = {
  id: '123',
  status: 'active',
  user: { email: 'admin@corp.com' },
  started_at: '2026-01-15T10:00:00Z',
};

describe('SessionHeader', () => {
  it('passes session status to StatusIndicator', () => {
    render(SessionHeader, { props: { session: mockSession } });
    // StatusIndicator renders the status — verify it's visible
    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('emits onaction when action button clicked', async () => {
    const onaction = vi.fn();
    render(SessionHeader, { props: { session: mockSession, onaction } });
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: /approve/i }));
    expect(onaction).toHaveBeenCalledWith('approve');
  });
});
```

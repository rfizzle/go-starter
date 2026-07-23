# Storybook Conventions (Svelte 5 + Tailwind 4)

## Story file location
Stories live alongside the component they document:
```
web/src/lib/components/<PascalName>/<PascalName>.stories.ts
web/src/lib/compositions/<PascalName>/<PascalName>.stories.ts
```

## Story file structure (CSF3 format)

```ts
import type { Meta, StoryObj } from '@storybook/svelte';
import MyComponent from './MyComponent.svelte';

const meta = {
  title: 'Components/MyComponent',   // or 'Compositions/MyComponent'
  component: MyComponent,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['primary', 'secondary', 'ghost', 'danger'],
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
    disabled: { control: 'boolean' },
    loading: { control: 'boolean' },
  },
} satisfies Meta<typeof MyComponent>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    variant: 'primary',
    size: 'md',
  },
};

export const AllVariants: Story = {
  // Use render function for multi-instance stories
};
```

## Story hierarchy
Organize under these top-level groups:

| Group | What goes here |
|-------|---------------|
| `Tokens/` | Color palette, typography scale, spacing, motion demos |
| `Components/` | Atomic components (Button, Badge, StatusIndicator) |
| `Compositions/` | Multi-component patterns (SessionHeader, FilterPanel) |
| `Views/` | Full page mockups (for visual regression / design review) |

## Required stories per component

### Atomic components
| Story | Purpose |
|-------|---------|
| `Default` | Component with default props |
| `AllVariants` | Grid showing every variant × size combination |
| `States` | All interactive states: hover, focus, active, disabled, loading |
| `DarkMode` | Pinned to dark via `globals: { theme: 'dark' }` |
| `WithLongContent` | Stress test with overflowing text |

### Compositions
| Story | Purpose |
|-------|---------|
| `Default` | Realistic production data |
| `Loading` | Skeleton/loading state |
| `Empty` | Empty state (no data) |
| `Error` | Error state |
| `Interactive` | Demonstrates key user flow with actions logged |

## Args and controls

Map every prop to a Storybook control:
- `string` enums → `control: 'select'` with `options`
- `boolean` → `control: 'boolean'`
- `number` → `control: 'number'` (with min/max/step if bounded)
- `string` (free text) → `control: 'text'`
- Complex objects → `control: 'object'` or omit from controls and set in args directly
- Callbacks → use `fn()` from `@storybook/test`

## Decorators

### Dark mode stories

Dark mode is handled globally via `.storybook/preview.ts`. A toolbar toggle
lets reviewers switch themes interactively, and per-story globals pin a story
to dark mode for snapshot/regression purposes.

```ts
export const DarkMode: Story = {
  args: {},
  globals: { theme: 'dark' },
};
```

The global decorator in `preview.ts` reads `globals.theme` and applies the
`.dark` class to `document.body` + `document.documentElement`, which activates
the dark token overrides defined in `app.css`.

### Layout decorators
```ts
// Centered content
parameters: { layout: 'centered' }

// Full-width (for page-level compositions)
parameters: { layout: 'fullscreen' }

// Padded (default)
parameters: { layout: 'padded' }
```

## Interaction tests (play functions)

For compositions with interactive flows, use play functions.

**Required package:** `@storybook/test` (installed as a dev dependency). This package
provides `within`, `userEvent`, `expect`, and `fn` for story interactions and action
logging. Do NOT use the legacy `@storybook/addon-actions` — it is empty in Storybook 8+.

```ts
import { within, userEvent, expect, fn } from '@storybook/test';

export const Interactive: Story = {
  args: {
    onAction: fn(),  // replaces the deprecated action() from addon-actions
    /* realistic data */
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);

    // Simulate user flow
    await userEvent.click(canvas.getByRole('button', { name: /filter/i }));
    await expect(canvas.getByRole('listbox')).toBeVisible();

    await userEvent.click(canvas.getByRole('option', { name: /active/i }));
    await expect(canvas.getByText('1 filter applied')).toBeVisible();
  },
};
```

## Accessibility audit

If the host project does not configure `@storybook/addon-a11y`, run the separate
`pnpm --dir web axe-scan` gate against a live Storybook instance and report an unavailable
browser/runtime as unverified. Do not claim automatic per-story auditing unless the addon
is installed and configured in `.storybook/main.ts`.

## Documentation (autodocs)

With `tags: ['autodocs']`, Storybook generates a docs page. Enhance it:
- Write a JSDoc comment on the component's `$props()` type for prop descriptions
- Add a `docs.description.component` in meta parameters for the component overview
- Use named stories that read as a feature tour (Default → Variants → States → Edge cases)

## Realistic data

For compositions, always use data shaped like production:
```ts
const mockSession = {
  id: 'sess_01HX7...',
  status: 'active',
  risk_score: 73,
  user: { email: 'alice.chen@corp.com', name: 'Alice Chen' },
  target: { host: 'prod-db-primary.internal', protocol: 'ssh' },
  started_at: '2026-06-15T14:32:00Z',
  duration_seconds: 1847,
  stream_count: 3,
};
```

Never use placeholder data like "Lorem ipsum", "test@test.com", or sequential IDs.
Use realistic but fictional values that exercise formatting (long emails, multi-word
names, various protocols, edge-case timestamps).

# Bits UI Testing Conventions (jsdom / Vitest)

Bits UI primitives use pointer events and DOM APIs that jsdom does not implement.
This reference documents the required polyfills and interaction patterns for testing
Bits UI-backed components in a jsdom environment with `@testing-library/svelte`.

## Required: Pointer Capture Polyfill

Bits UI Select (and other pointer-event-driven primitives) calls
`element.hasPointerCapture()` and `element.releasePointerCapture()` internally.
jsdom does not implement the Pointer Capture API, causing unhandled `TypeError`
exceptions during tests.

**Add to `src/tests/setup.ts`:**

```ts
if (!HTMLElement.prototype.hasPointerCapture) {
  HTMLElement.prototype.hasPointerCapture = () => false;
}
if (!HTMLElement.prototype.releasePointerCapture) {
  HTMLElement.prototype.releasePointerCapture = () => {};
}
if (!HTMLElement.prototype.setPointerCapture) {
  HTMLElement.prototype.setPointerCapture = () => {};
}
```

Without this polyfill, tests will report "Uncaught Exception: TypeError:
target?.hasPointerCapture is not a function" even though assertions may still pass.

## Opening Bits UI Primitives

### Select / Combobox

Bits UI Select triggers open on `pointerdown`, **not** `click`.

```ts
// ✓ Correct — opens the dropdown
await fireEvent.pointerDown(trigger, { pointerType: 'mouse' });

// ✗ WRONG — does NOT open the dropdown
await fireEvent.click(trigger);
```

### DropdownMenu / Dialog

These open on `click`:

```ts
// ✓ Correct
await fireEvent.click(trigger);
```

## Selecting Items

### Select.Item

Bits UI Select items handle selection on `pointerup` (for mouse interactions):

```ts
// ✓ Correct — selects the item
const option = screen.getByRole('option', { name: 'SSH' });
await fireEvent.pointerUp(option, { pointerType: 'mouse' });

// ✗ WRONG — does not trigger selection
await fireEvent.click(option);
```

### DropdownMenu.Item

Menu items respond to `click`:

```ts
// ✓ Correct
await fireEvent.click(screen.getByText('Sign out'));
```

## ARIA Roles

| Bits UI Primitive | Trigger role | Content role | Item role |
|-------------------|-------------|--------------|-----------|
| Select | `button` (with `aria-haspopup="listbox"`, `aria-expanded`) | `listbox` | `option` |
| DropdownMenu | `button` (with `aria-haspopup="menu"`, `aria-expanded`) | `menu` | `menuitem` |
| Dialog | (varies) | `dialog` | — |
| Combobox | `combobox` (with `aria-expanded`) | `listbox` | `option` |

**Key:** Query triggers by `getByRole('button', { name: /Label/i })` for Select,
not `getByRole('combobox')` — that role is only used by the Combobox primitive.

## Portal Rendering

Bits UI renders content through a Portal (appended to `document.body`).
Testing Library's `screen.*` queries search the full document, so portal-rendered
content IS findable. No special container scoping is needed.

## `userEvent` vs `fireEvent`

Prefer `fireEvent` from `@testing-library/svelte` for Bits UI interactions.
`userEvent` dispatches a full sequence of pointer events that can conflict with
Bits UI's internal event handling (double-open, race conditions with jsdom
limitations).

For non-Bits-UI elements (text inputs, standard buttons), `userEvent` is still
preferred as it more closely simulates real browser behavior.

## Complete Example: Testing a Select Filter

```ts
import { render, screen, fireEvent } from '@testing-library/svelte';

it('selecting an option calls the change handler', async () => {
  const onchange = vi.fn();
  render(MyComponent, { props: { onchange } });

  // Open the select
  const trigger = screen.getByRole('button', { name: /Protocol/i });
  await fireEvent.pointerDown(trigger, { pointerType: 'mouse' });

  // Verify listbox appeared
  expect(screen.getByRole('listbox')).toBeInTheDocument();

  // Select an option
  const option = screen.getByRole('option', { name: 'SSH' });
  await fireEvent.pointerUp(option, { pointerType: 'mouse' });

  // Verify callback
  expect(onchange).toHaveBeenCalledWith('ssh');
});
```

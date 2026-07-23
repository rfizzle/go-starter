# `<CompositionName>` Composition Spec

> Status: draft · Spec version: 2 · Owner: · Feature:

## Purpose

What the composition does, where it appears, and when to keep the layout inline instead.

## Children

| Component | Source | Props passed | Events consumed |
|---|---|---|---|

## Props API

```ts
type Props = {
  // typed data, callbacks, snippets, and class override
};
```

## Internal State

State the composition owns locally and state that remains owned by the caller.

## Interaction Flows

| User action | Internal effect | Emitted event or navigation |
|---|---|---|

## Layout

Describe semantic structure, responsive behavior, and configured token/scale usage.

## Accessibility Delegation

Describe reading order, focus sequence, landmarks, announcements, and which child owns
each widget-level behavior.

## Testing Scenarios

- [ ] Data reaches each child correctly.
- [ ] Callback payloads and state transitions match the contract.
- [ ] Loading, empty, and error behavior is covered where applicable.
- [ ] Focus and reading order remain logical at supported layouts.

## Open Questions

List unresolved decisions and the person or artifact that can resolve each one, or
write `None`.

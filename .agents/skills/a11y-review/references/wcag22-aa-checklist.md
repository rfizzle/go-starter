# WCAG 2.2 AA Checklist and APG Routing

> Baseline: WCAG 2.2 Level A and AA. Treat AAA criteria as advisory unless the product
> explicitly adopts them. Verify version-sensitive behavior against the linked W3C
> sources before approving a new interaction pattern.

## Perceivable

- [ ] Normal text contrast is at least 4.5:1; large text is at least 3:1.
- [ ] Required visual boundaries, state indicators, icons, and focus indicators meet
      non-text contrast requirements where WCAG 1.4.11 applies.
- [ ] Color is not the only way information or state is conveyed.
- [ ] Content reflows at 320 CSS px without losing information or operation.
- [ ] Text can be resized to 200% without clipping or loss.
- [ ] Motion respects `prefers-reduced-motion` and avoids vestibular triggers.

## Operable

- [ ] Every operation is keyboard-accessible and no keyboard trap exists.
- [ ] Focus is always visible and is not fully obscured by author-created content
      (WCAG 2.4.7 and 2.4.11, AA).
- [ ] Pointer targets meet WCAG 2.5.8 or a documented exception.
- [ ] Hover/focus-revealed content is dismissible, hoverable, and persistent where
      WCAG 1.4.13 applies.
- [ ] Dragging has a non-drag alternative where WCAG 2.5.7 applies.
- [ ] Focus order follows meaning and operation.

The enhanced focus-not-obscured criterion (2.4.12) and Focus Appearance (2.4.13) are
Level AAA. They are useful design goals, but do not report them as AA failures.

## Understandable

- [ ] The accessible name contains the visible label (WCAG 2.5.3).
- [ ] Errors are identified in text, associated with fields, and include a useful
      correction suggestion when known.
- [ ] State changes are announced when they are not otherwise apparent.
- [ ] Consistent help, redundant entry, and accessible authentication criteria are
      evaluated when the component participates in those flows.

## Robust

- [ ] Native HTML semantics are used before ARIA.
- [ ] Name, role, state, value, and relationships are correct for every control.
- [ ] IDs and ARIA references remain unique when multiple instances render.
- [ ] No Svelte accessibility warning is suppressed without a reviewed justification.

## Required manual evidence

- [ ] Complete every interaction using the keyboard only.
- [ ] Smoke-test with at least one screen-reader/browser combination supported by the
      product; record combinations not available in the current environment as unverified.
- [ ] Toggle the operating system's reduced-motion preference.
- [ ] Verify RTL only when the product claims RTL support or the component contract
      explicitly requires it.

## APG pattern routing

Do not copy a condensed row as the component's full keyboard contract. Open the governing
APG page, distinguish required from optional keys, and summarize the exact model needed by
the component. Record the source URL in the spec.

| Component | Governing source | Important review note |
|---|---|---|
| Button | [Button pattern](https://www.w3.org/WAI/ARIA/apg/patterns/button/) | Native `button` supports Enter and Space. |
| Modal dialog | [Dialog modal pattern](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/) | Initial focus, contained Tab sequence, Escape, and focus return are contract decisions. |
| Disclosure | [Disclosure pattern](https://www.w3.org/WAI/ARIA/apg/patterns/disclosure/) | Prefer native `button`; Enter and Space toggle. |
| Accordion | [Accordion pattern](https://www.w3.org/WAI/ARIA/apg/patterns/accordion/) | Separate required keys from optional arrow/Home/End navigation. |
| Tabs | [Tabs pattern](https://www.w3.org/WAI/ARIA/apg/patterns/tabs/) | Specify orientation and automatic versus manual activation. |
| Menu button | [Menu button pattern](https://www.w3.org/WAI/ARIA/apg/patterns/menu-button/) | Button-opening keys and menu navigation are separate parts of the model. |
| Listbox | [Listbox pattern](https://www.w3.org/WAI/ARIA/apg/patterns/listbox/) | Specify single/multi-select behavior, selection model, and type-ahead policy. |
| Combobox | [Combobox pattern](https://www.w3.org/WAI/ARIA/apg/patterns/combobox/) | Behavior varies by editable/select-only and autocomplete mode. |
| Switch | [Switch pattern](https://www.w3.org/WAI/ARIA/apg/patterns/switch/) | Space is required; Enter is optional. Use `aria-checked`. |
| Checkbox | [Checkbox pattern](https://www.w3.org/WAI/ARIA/apg/patterns/checkbox/) | Space toggles; define mixed-state behavior when applicable. |
| Radio group | [Radio group pattern](https://www.w3.org/WAI/ARIA/apg/patterns/radio/) | Arrow behavior differs when the group is inside a toolbar. |
| Slider | [Slider pattern](https://www.w3.org/WAI/ARIA/apg/patterns/slider/) | Touch assistive technology support requires manual verification. |
| Tooltip | [Tooltip pattern](https://www.w3.org/WAI/ARIA/apg/patterns/tooltip/) | Focus remains on the trigger and Escape dismisses the tooltip. |

Primary references:

- [WCAG 2.2](https://www.w3.org/TR/WCAG22/)
- [WCAG 2.2 quick reference](https://www.w3.org/WAI/WCAG22/quickref/)
- [ARIA Authoring Practices Guide](https://www.w3.org/WAI/ARIA/apg/patterns/)

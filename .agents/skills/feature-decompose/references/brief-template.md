# Feature Brief: `<feature-name>`

> Status: draft · Owner: · Target route: · Date:
>
> Status vocabulary is exactly `draft | approved`. Downstream component/composition
> builds are blocked by the `check-specs` gate until each unit's own spec is `approved`.

## 1. UX Requirement
What does the user need to accomplish? One paragraph, user-centric language.
Include the user role, the goal, and what success looks like.

## 2. Token Dependencies
Tokens this feature requires. Flag any that are missing from the current `@theme` block.

| Token | Purpose | Status |
|-------|---------|--------|
| `bg-surface` | Page background | Exists |
| `color-warning` | Overdue indicator | **Missing — request from design-tokens** |

## 3. Component Inventory

Default every new unit to generic (parameterized, reusable). Classify a unit as
specific only when the interaction is high-value or extreme enough that generalizing it
would visibly harm the target UX, and record why in the "Why specific" column.

### Existing (reuse as-is)
| Component | Path | Used for |
|-----------|------|----------|
| `StatusIndicator` | `components/display/StatusIndicator.svelte` | Session status badges |

### Needs Enhancement
| Component | Path | Change needed |
|-----------|------|---------------|
| `DataTable` | `components/display/DataTable.svelte` | Add row-expansion support |

### New Generic Components (atomic — `component-spec` → `a11y-review` → `component-build`)
| Component | Brief | APG pattern | Backing primitive |
|-----------|-------|-------------|-------------------|
| `CommandPreview` | Inline code block with copy button | none (static) | none |

### New Specific Components (atomic, one-off — same pipeline)
| Component | Brief | APG pattern | Backing primitive | Why specific |
|-----------|-------|-------------|-------------------|---------------|

### New Generic Compositions (multi-component, reusable — `composition-spec` → `composition-build`)
| Composition | Brief | Children |
|-------------|-------|----------|
| `SessionHeader` | Page header with metadata chips and actions | StatusIndicator, UserChip, TimeAgo, Button |

### New Specific Compositions (multi-component, one-off — same pipeline)
| Composition | Brief | Children | Why specific |
|-------------|-------|----------|---------------|

## 4. Data Requirements
| Source | Endpoint / Store | Shape | Notes |
|--------|-----------------|-------|-------|
| API | `GET /api/v1/sessions/:id` | `SessionDetail` | Includes nested streams |
| Derived | client-side | `activeTab: string` | URL param sync |

## 5. Interaction Map
Describe every user-initiated action and its effect:

| Action | Trigger | Effect | State change |
|--------|---------|--------|--------------|
| Click stream row | row click | Navigate to stream detail | URL push |
| Toggle tab | tab click | Show different content panel | `activeTab` update |
| Approve session | button click | API call → status change | Optimistic update + refetch |

## 6. State Management
- What state lives in the URL (bookmarkable)?
- What state is component-local?
- What state needs to be shared across compositions?

## 7. Testing Scenarios

### Golden path
- [ ] Page loads with valid session data, all sections populated
- [ ] User can navigate between tabs

### Edge cases
- [ ] Session with no streams (empty state)
- [ ] Extremely long user email (truncation)
- [ ] Session still active (duration shows "ongoing", status pulses)
- [ ] API error on load (error state with retry)
- [ ] Deep link to specific tab via URL

### Accessibility
- [ ] All interactive elements keyboard-reachable
- [ ] Tab panel focus management
- [ ] Screen reader announces tab changes

## 8. Build Order
Ordered list respecting dependencies:

1. **Tokens** — request any missing tokens from `design-tokens`
2. **Components** — build new atomic components (parallel where independent)
3. **Compositions** — build multi-component patterns (depends on step 2)
4. **View assembly** — wire compositions into the page route
5. **E2E test** — golden path integration test

## 9. Open Questions
Decisions deferred to review. Each should name who can unblock it.

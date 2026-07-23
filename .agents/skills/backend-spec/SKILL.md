---
name: backend-spec
description: "Author and review contract-first Go backend module and cross-module flow specs. Use before implementing, reimplementing, or behaviorally refactoring a business module; when documenting exported contracts, invariants, errors, collaborators, recovery, and tests; or when checking code against its approved spec. Follow the host repository's naming and placement conventions. Not for frontend specs or file-placement decisions."
license: MIT
---


# Backend Spec

Spec-driven development for Go backends. A business package can carry a **module spec**
that states, in functional terms, what the package does and the interface it adheres to,
so the package can be **reimplemented from the spec alone**. Cross-module capabilities
use a **flow spec**.

> "Clear is better than clever." — Go Proverbs
>
> The spec is the contract. The code is one valid implementation of it.

To deliberately deviate from this pattern for a package, add a short note in that package's spec explaining why.

## Why this exists

Go backends benefit from two levels of behavioral contract:

| Backend tier | Typical file | Scope |
|---|---|---|
| **Module spec** | `<pkg>.spec.md` near the package | One Go package. A leaf module and an orchestrator share one template; the orchestrator gives more weight to **Collaborators**. |
| **Flow spec** | `<name>.flow.spec.md` in the repository's documentation area | A cross-module capability that no single package owns. |

## When this applies

- Creating a **new business package** → approve its module spec before writing
  implementation code.
- **Reimplementing or refactoring** an existing package → the spec is the source of truth for what must still hold; update it in the same change if behavior changes.
- Documenting a **capability that spans packages** → write a flow spec.
- **Reviewing** a package → check the code still satisfies its spec (contract, invariants, error contract).

Platform packages may carry a focused technical contract when explicitly requested, but
they are not part of the mandatory business-module spec inventory by default.

This skill is **not** about where a file lives (see `go-project-layout`), naming identifiers (see `golang-naming`), or in-package design patterns (see `golang-design-patterns`). Reference those from the spec; don't restate them.

## Core principles

1. **Functional, not implementation.** The spec describes *what* and *the contract*, never the SQL, the exact struct layout, or the query plan. If a reader could rebuild the package a different way and still satisfy every line of the spec, the spec is at the right altitude.
2. **The Contract section is the anchor.** It is the exported surface — the `domain` interface the package satisfies, the constructor, and any exported functions/types — copied as Go signatures with one-line intent. This is the "one spot to understand the interface it adheres to."
3. **KISS is enforced by omission.** If a behavior isn't in the spec, the implementation shouldn't have it. A spec that grows a dozen edge-case rules is a signal the package is doing too much — split it.
4. **Invariants are the reimplementation gold.** Section 4 (Behaviors & Invariants) captures the non-obvious rules a fresh implementation would get wrong: idempotency, ordering guarantees, what a state transition means, dedup keys, banding thresholds. Write these as testable statements.
5. **One spec per package, co-located.** It sits next to the code it governs, exactly like the frontend co-locates component specs.
6. **Status header, like the frontend.** Every spec opens with a one-line status blockquote so its maturity is visible at a glance.

---

## Template A — Module Spec

Filename: `<package>.spec.md`, in or next to the package according to repository
conventions (for example, `internal/orders/orders.spec.md`).

````markdown
# `<package>` Module Spec

> Status: draft | approved · Layer: modules · Owner: backend · Depends on: <domain, platform/database, sibling modules — or "none">

## 1. Purpose
One paragraph. The single responsibility of this package and where it sits in the
system. What it owns; what it explicitly does NOT do. If it satisfies a
`domain` interface, name it here.

## 2. Contract
The exported surface, as Go signatures with a one-line intent each. This is the
interface the package adheres to — copy the real signatures.

**Satisfies:** `domain.<Interface>` (compile-time checked with `var _ ... = (*T)(nil)`)

**Constructor:**
```go
func New<Type>(deps...) *<Type>   // what it wires
```

**Methods / exported functions:**
```go
func (r *Repository) GetByID(ctx, id) (*domain.X, error)   // one-line intent + error behavior
func Enqueue<Thing>(ctx, enq, args) (bool, error)          // pure/free functions too
```

List only the exported surface. Internal helpers are implementation detail —
leave them out.

## 3. Data Model
Tables/columns owned (name + role, not full DDL — the migration is the DDL source
of truth). Key `domain` types read or written. Object-store keys or queues touched.
For a pure engine with no persistence, write "None — pure/in-memory."

## 4. Behaviors & Invariants
The rules a reimplementation MUST preserve, as testable statements. This is the
most important section. Examples of the *kind* of rule that belongs here:
- Idempotency: "Retrying the same command does not create a duplicate record."
- Ordering: "Results are returned by `created_at ASC`, then ID for ties."
- Derivation: "Orders above the configured approval threshold require review."
- Dedup: "The active-job index coalesces NULL stream_id to the nil UUID."
- Truncation/limits: "GetDetail returns at most the top 3 signals by score."
Number them so tests and reviews can cite them (I1, I2, ...).

## 5. Error Contract
Which sentinel errors this package returns and when (`database.ErrNotFound`,
domain sentinels). How it wraps (`fmt.Errorf("<pkg>.<Method>: %w", err)`).
Which failure modes map to which error. What is retryable vs terminal (for
workers). Reference `golang-error-handling` rather than restating it.

## 6. Collaborators
The dependency edges. For a leaf module this is short (domain + platform). For an
orchestrator, this is where the "composition" lives — list each sibling module it
coordinates and the seam (interface) it depends on, not the concrete type.

| Dependency | Kind | Used for |
|---|---|---|
| `domain` | contract | models + repository interface |
| `platform/database` | infra | Querier, ScanOne/All, WhereBuilder, sentinels |
| `modules/notifications` | sibling | emits completion events (orchestrators only) |

State that `domain` and `platform` remain independent leaves: modules may depend on each,
neither leaf depends on modules, and modules never import `internal/api/ogen`.

## 7. Testing Scenarios
A checklist mirroring the frontend spec's testing section. Split by kind.
Each invariant in §4 should have at least one line here.

### Unit (pgxmock / pure)
- [ ] ...
### Integration (real PostgreSQL via Docker Compose)
- [ ] ...

## 8. Open Questions
Unresolved decisions, or "None."
````

### Leaf vs orchestrator

Same template, different weight:
- **Leaf module** (`accounts`, `orders`, `rules`): §2 Contract and §4 Invariants are the
  bulk; §6 Collaborators is a couple of rows (domain + platform).
- **Orchestrator module** (`checkout`, `fulfillment`): §6 Collaborators is substantial—it
  names every sibling module and the narrow interface it depends on. §4 captures
  coordination rules such as job dispatch, deduplication, and phase ordering. The
  orchestrator's own contract is often a worker or enqueuer seam, not a repository.

---

## Template B — Flow Spec

Filename: `<name>.flow.spec.md` in the repository's documentation area. Use it for a
capability that spans packages and has no single owner.

````markdown
# Flow Spec: `<name>`

> Status: draft | approved · Owner: backend · Spans: <modules involved> · Date: YYYY-MM-DD

## 1. Capability
The functional outcome, in one paragraph. What the system does end-to-end and why
it matters. Written so a reader who knows none of the packages understands the goal.
Include the success criterion.

## 2. Participants
The packages that collaborate and each one's role in this flow. Link to each
module spec.

| Package | Role in this flow | Module spec |
|---|---|---|
| `orders` | ... | `internal/orders/orders.spec.md` |

## 3. Sequence
The end-to-end path as ordered steps or a diagram. Where each step happens
(which package), what triggers it, what it produces. Note sync vs async /
scheduled vs real-time boundaries.

## 4. Data & State Transitions
The state machine or data lifecycle this flow drives (e.g. pending → in_review →
approved). Where state is persisted and which package owns each transition.

## 5. Invariants & Guarantees
Cross-cutting rules that only make sense at the flow level: exactly-once vs
at-least-once, dedup keys shared across packages, idempotency of the whole path,
ordering guarantees between phases.

## 6. Failure & Recovery
What happens when a step fails: retry/backoff, dead-lettering, reconciliation,
partial-progress recovery. What the system guarantees after a crash mid-flow.

## 7. Testing Scenarios
End-to-end scenarios (often integration/e2e), plus the seams where a flow-level
test injects fakes.
- [ ] Golden path ...
- [ ] Failure/recovery ...

## 8. Open Questions
Or "None."
````

---

## Authoring workflow

1. **Read the code (or the intent, for new work).** For an existing package, read the `domain` interface it satisfies, the constructor, and the exported methods. For new work, the interface may not exist yet — the spec drives it.
2. **Fill Contract first.** Copy the real Go signatures. If they don't exist yet, this is where you design them — keep the surface minimal.
3. **Mine the invariants.** Read the method bodies (or design them) for the non-obvious rules and record them as numbered, testable statements in §4. This is where most reimplementation value lives.
4. **Keep altitude.** Delete anything that describes *how* rather than *what*. No SQL, no query plans, no struct field lists beyond what the contract needs.
5. **Cross-link.** Reference sibling module specs and the relevant `golang-*` skills instead of restating their rules.
6. **Verify.** Check every signature against the code or intended API. For an existing
   package, confirm each §4 invariant has a corresponding scenario in §7.
7. **Request approval.** Present the draft and stop. Do not mark it `approved`, write Go
   implementation, or treat silence as approval. Continue only after explicit human
   approval; if approval changes the contract, update and re-present the draft.

## Quality bar (self-review before requesting approval)

- [ ] Could someone delete the package's `.go` files and rebuild a *behaviorally equivalent* package from this spec alone? If not, §2/§4/§5 are incomplete.
- [ ] Is there any SQL, struct-field dump, or query detail that isn't load-bearing? Remove it.
- [ ] Does every exported symbol appear in §2, and does §2 contain nothing unexported?
- [ ] Is every §4 invariant a *testable* statement, and does it appear in §7?
- [ ] Does §6 describe the actual leaf dependency graph and list real seams (interfaces),
      not concrete clients used by business logic?
- [ ] For orchestrators: is every coordinated sibling in §6, or is a flow spec the better home?

## Anti-patterns

- **Restating the code.** A spec that is pseudo-code of the implementation adds no value and rots instantly. Describe the contract and the rules, not the steps.
- **SQL in the spec.** The migration and `repository.go` own the SQL. The spec owns "returns not-found as `database.ErrNotFound`," not the `SELECT`.
- **One giant spec for a module tree.** Each independently reimplementable sub-package
  gets its own module spec; the parent orchestration package gets a spec that references
  them.
- **Flow spec for a single package.** If one package owns the whole thing, it's a module spec, not a flow.
- **Letting the spec drift.** If a change alters a contract, an invariant, or an error behavior, update the spec in the same commit. A stale spec is worse than none.

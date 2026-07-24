# Implementation Plan: [FEATURE_NAME]

**Source:** `docs/features/[YYYY-MM-DD-feature-name.md]`  
**Date:** [YYYY-MM-DD]  
**Status:** [Planning/Ready/In Progress/Complete]

<!-- Replace placeholder text and remove instructions that do not apply. -->

## Outcome and Boundary

- **Problem:** [Problem being solved]
- **Required outcome:** [Observable result]
- **In scope:** [Acceptance criteria covered by this plan]
- **Non-goals:** [Related behavior intentionally excluded]
- **Existing behavior preserved:** [Compatibility requirements]

Stop planning when every acceptance criterion has an implementation and verification
path. Do not add adjacent improvements to make the plan appear comprehensive.

## Existing System

- **Relevant packages:** [Current code that owns the behavior]
- **Existing contracts:** [Exported APIs, interfaces, OpenAPI operations, or database
  contracts already available]
- **Existing implementation pattern:** [Pattern to reuse]
- **Constraints:** [Known bounds or compatibility requirements]

Prefer modifying an existing clear owner over creating a new layer or parallel path.

## Simplest Viable Design

Describe the least complex design that satisfies the approved scope:

[Short implementation description]

Explain why a smaller direct change is insufficient only when additional structure is
required.

### Package Contract Assessment

For every new or materially changed package:

| Package | Purpose and non-goals | Public contract | Implementation | Documentation |
| --- | --- | --- | --- | --- |
| `[path]` | [One responsibility; explicit exclusions] | [Concrete API or justified interface] | [Local or named subpackages] | [`doc.go`; README if required] |

Apply these rules:

- A documented contract does not require an interface.
- Prefer consumer-defined narrow interfaces for dependency seams.
- Keep one natural implementation local to the contract package.
- Put multiple real variants in named subpackages beneath a shared contract.
- Return concrete types from constructors unless the contract requires otherwise.
- Add compile-time assertions for concrete interface implementations.
- Require a concise README for business modules, substantial contracts, multiple
  implementations, or important lifecycle and operational behavior.

### Import and Ownership Boundaries

- **Contract owner:** [Package that defines stable shared behavior]
- **Implementation owner:** [Package or subpackages containing concrete behavior]
- **Composition root:** [Where concrete dependencies are constructed and injected]
- **Forbidden dependencies:** [Imports or implementation details callers must not use]

## Contract Changes

Include only applicable changes:

- **OpenAPI:** [Operation and schema changes or `None`]
- **Go API:** [Exported type, function, method, or interface changes or `None`]
- **Persistence:** [Migration, query, or transaction changes or `None`]
- **Configuration:** [Required configuration change or `None`]
- **Errors:** [Stable error outcomes and translation or `None`]

For API changes, update `api/openapi.yml` first and regenerate
`internal/api/v1/oas_*_gen.go`. Never edit generated API files directly.

## Operational Bounds

- **Input and result limits:** [Bounds and rejection behavior]
- **Timeout and cancellation:** [Propagation and ownership]
- **Concurrency and resource limits:** [Bounds or `Not applicable`]
- **Retry, fallback, and degradation:** [Specified bounded behavior or `None`]
- **Cleanup and partial failure:** [Resource and atomicity behavior]

Do not introduce an unbounded mechanism or a resilience mechanism without a documented
failure mode.

## Complexity Gate

For each selected item, record the current concrete need and approval status:

| Addition | Needed? | Justification | Approval |
| --- | --- | --- | --- |
| External dependency | [Yes/No] | [Current need] | [Approved/Not required] |
| New package or layer | [Yes/No] | [Current need] | [Approved/Not required] |
| One-implementation interface | [Yes/No] | [Why concrete API is insufficient] | [Approved/Not required] |
| Persistent state or configuration | [Yes/No] | [Current need] | [Approved/Not required] |
| Concurrency or background work | [Yes/No] | [Current need and bound] | [Approved/Not required] |
| Cache, retry, fallback, or queue | [Yes/No] | [Failure mode and bound] | [Approved/Not required] |
| Reusable abstraction | [Yes/No] | [Current consumers] | [Approved/Not required] |

Remove any unjustified addition from the design. Material scope or maintenance expansion
must be approved before implementation.

## Implementation Steps

Keep steps small, ordered, and traceable to approved behavior.

1. [Contract or specification change required by AC#]
2. [Smallest implementation change required by AC#]
3. [Boundary or implementation wiring required by AC#]
4. [Documentation and verification]

Do not create placeholder layers, speculative extension points, or unrelated refactoring.

## Verification Plan

Choose the narrowest effective verification for each requirement or significant risk.

| Requirement or risk | Level | Check | Defect it catches |
| --- | --- | --- | --- |
| AC1 | [Unit/package/integration/E2E/inspection] | [Test or command] | [Realistic defect] |
| AC2 | [Level] | [Test or command] | [Realistic defect] |
| [Boundary or regression] | [Level] | [Test or command] | [Realistic defect] |

- Do not require unit, integration, and E2E tests for every change.
- Reuse one contract test suite when multiple implementations share meaningful behavior.
- Do not test generated code, trivial accessors, library behavior, or private call order.
- Coverage percentage is not a completion target.
- Record intentionally omitted tests when the omission is relevant to review.

## Required Commands

Run the smallest relevant set during implementation and the repository checks appropriate
to the final diff:

```bash
make generate
make fmt
make check
```

Use `make race`, `make test-integration`, `make test-all`, vulnerability
scanning, or benchmarks only when the
changed risk warrants them.

## Definition of Done

- [ ] Every implementation step maps to approved scope.
- [ ] Every acceptance criterion is satisfied and verified.
- [ ] No material behavior was added beyond the specification.
- [ ] Package contracts, bounds, and non-goals are documented.
- [ ] Interfaces and implementation packages have current concrete justification.
- [ ] Meaningful tests pass without redundant risk coverage.
- [ ] Generated artifacts and relevant documentation are current.
- [ ] Applicable formatting, linting, build, and test checks pass.
- [ ] Unsupported behavior and intentionally omitted tests are recorded.

## Plan Validation

- [ ] The plan identifies the smallest viable design.
- [ ] Existing contracts and patterns are reused where suitable.
- [ ] New complexity passes the complexity gate.
- [ ] Package ownership and import direction are explicit.
- [ ] One implementation stays local; multiple variants use named subpackages.
- [ ] Verification maps to contracts and risks rather than coverage targets.
- [ ] Non-goals prevent likely scope expansion.

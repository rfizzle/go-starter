# Feature: [FEATURE_NAME]

<!-- Replace placeholder text and remove instructions that do not apply. -->

## Metadata

**ID:** [YYYY-MM-DD-feature-name]  
**Created:** [YYYY-MM-DD]  
**Status:** [Draft/Review/Approved/In Development/Complete]  
**Owner:** [Name/Role]

## User Need

**As a** [user type]

**I want** [specific capability]

**So that** [observable benefit]

## Problem

**Current behavior:** [What happens today]

**Desired outcome:** [What must become possible]

## Scope

### In Scope

- [Smallest capability required to produce the desired outcome]
- [Required boundary, error, or compatibility behavior]

### Non-Goals

- [Related behavior this feature intentionally does not provide]
- [Speculative extension, configuration, or optimization that is excluded]

Do not treat non-goals as deferred implementation work unless a later specification
explicitly adds them.

## Public Contract

Describe only externally observable behavior:

- **Inputs:** [Accepted inputs and their meaning]
- **Outputs:** [Results and externally visible effects]
- **Errors:** [Stable error outcomes callers may rely on]
- **Compatibility:** [Existing behavior that must remain unchanged]
- **API:** [OpenAPI operations or `None`]
- **Persistence:** [Required durable state or `None`]

If the feature introduces or materially changes a package, identify:

- **Package purpose:** [One clear responsibility]
- **Package non-goals:** [Responsibilities kept outside the package]
- **Contract form:** [Exported concrete API, shared interface, or consumer-local interface]
- **Implementation shape:** [One local implementation or named implementation subpackages]
- **Documentation:** [Package documentation and whether a README is required]

Do not require an interface when a documented concrete API is sufficient.

## Operational Bounds

Specify only bounds relevant to this feature:

- **Input and payload limits:** [Limit and rejection behavior]
- **Time and deadlines:** [Timeout or cancellation behavior]
- **Pagination and result limits:** [Bounds or `Not applicable`]
- **Concurrency and resource limits:** [Bounds or `Not applicable`]
- **Retry or degradation policy:** [Bounded specified behavior or `None`]

No retry, fallback, cache, queue, or background worker is implied unless it is explicitly
specified here.

## Acceptance Criteria

- [ ] **AC1:** Given [context], when [action], then [observable result]
- [ ] **AC2:** Given [context], when [action], then [observable result]
- [ ] **AC3:** Given [context], when [action], then [observable result]

Add only criteria necessary to define the feature boundary. Each criterion must be
specific, measurable, and independently verifiable.

## Verification

Map each acceptance criterion or significant risk to the narrowest effective
verification. Do not require every test level.

| Requirement or risk | Verification | Defect it would catch |
| --- | --- | --- |
| AC1 | [Unit, package, integration, E2E, inspection, or command] | [Realistic failure] |
| AC2 | [Verification] | [Realistic failure] |
| [Named boundary or risk] | [Verification] | [Realistic failure] |

Tests without a meaningful contract, regression, boundary, or risk should not be added.
Coverage percentage is not a completion criterion.

## Complexity Assessment

Mark each item introduced by the proposed feature and justify it:

- [ ] New external dependency
- [ ] New package or architectural layer
- [ ] New interface with one expected implementation
- [ ] New persistent state or configuration
- [ ] Concurrency, background work, caching, retries, or queues
- [ ] Reusable abstraction without multiple current consumers
- [ ] Behavior not directly traceable to an acceptance criterion

**Justification or `None`:** [Why each selected item is necessary]

Material scope, operational, or maintenance expansion requires approval before
implementation.

## Dependencies and Risks

- **Blocked by:** [Required prerequisite or `None`]
- **External dependency:** [Required external system or `None`]
- **Material risk:** [Risk and bounded mitigation or `None`]

Do not invent risks or mitigations merely to fill this section.

## Definition of Done

- [ ] Acceptance criteria are met.
- [ ] Implemented behavior does not exceed approved scope.
- [ ] Public contracts and operational bounds are documented.
- [ ] Meaningful verification passes.
- [ ] Generated artifacts are current.
- [ ] Relevant formatting, linting, and test checks pass.
- [ ] Package documentation and README requirements are satisfied.
- [ ] Unsupported or intentionally omitted behavior is recorded.

## Sign-Off

- [ ] **Product owner or requester:** [Name] - [Date]
- [ ] **Technical reviewer:** [Name] - [Date]
- [ ] **Additional review:** [Security, operations, or `Not required`] - [Date]

## Template Validation

- [ ] All remaining content is specific to this feature.
- [ ] In-scope behavior and non-goals form a clear boundary.
- [ ] Acceptance criteria describe observable outcomes rather than implementation.
- [ ] Every planned test protects a named requirement or risk.
- [ ] Operational mechanisms are bounded and explicitly required.
- [ ] New abstractions and dependencies have current, concrete justification.
- [ ] Package contracts use the simplest suitable API and implementation shape.

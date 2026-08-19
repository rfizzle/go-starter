# AGENTS.md

This file provides guidance to AI Agents when working with code in this repository.

## Development Commands

### Essential Commands
- **Build**: `make build` - Build the local binary in `bin/`
- **Check**: `make check` - Run architecture, documentation, generation, formatting, vet, lint, and unit checks
- **Test**: `make test` - Run tests that do not require external services (alias for `make test-unit`)
- **Lint**: `make lint` - Run go vet and golangci-lint for code quality checks
- **Format**: `make fmt` - Format Go code with gofmt, goimports, and YAML files with yamlfmt
- **Generate**: `make generate` - Regenerate ogen and sqlc output from their sources
- **Development cycle**: `make generate fmt check` - Mutate generated/formatted sources, then run repository gates

### Additional Commands
- **Dependencies**: `make tidy` (rebuild go.mod and go.sum)
- **Environment check**: `make env` - Validate development environment setup
- **Race condition testing**: `make race` - Run the default Go suite with the race detector
- **Coverage**: `make cover` - Run both test tiers, write a coverage profile, and fail below the maintained-code floor (`MAINTAINED_COVERAGE_MIN` in `scripts/coverage-baseline.env`, currently 75%)
- **Vulnerability scan**: `make vuln` - Scan dependencies with govulncheck
- **Secret scan**: `make secrets-check` - Scan the full git history with gitleaks
- **Release snapshot**: `make release-snapshot` - Build the GoReleaser archive matrix
- **CI pipeline**: `make ci` - Run the read-only validation workflow (`check`, `secrets-check`, and `vuln`); see [docs/development/ci-contract.md](docs/development/ci-contract.md) for what each workflow enforces
- **Container image**: `make container` - Build the container image with Podman using build scripts
- **Skill validation**: `make skills-check` - Validate the skill library and its discovery links
- **Agent wiring**: `make agents-setup` - Regenerate the `.claude/` discovery symlinks

## Skills

Reusable task guidance lives in `.agents/skills/`, the canonical client-neutral
library described in [.agents/skills/README.md](.agents/skills/README.md). Each skill is a
directory holding a portable `SKILL.md`.

`.claude/skills/` and `.claude/agents/` contain **relative** symlinks back into
that library and into `docs/ai/prompts/`, so a clone gets working discovery
without copies that drift. `make agents-setup` regenerates them idempotently and
`make skills-check` validates the library and fails on a missing, misdirected, or
dangling link. Never copy a skill into `.claude/`; edit the canonical directory.

Route by task type; load the narrowest skill that covers the change:

| Task type | Skill | Verification |
| --- | --- | --- |
| Package placement, cross-cutting move, import direction | `go-project-layout`, `project-conventions` | `make architecture-check` |
| New or behaviorally changed business module | `backend-spec` | `make test` |
| Struct, interface, and constructor design | `golang-structs-interfaces`, `golang-design-patterns` | `make lint` |
| Naming, formatting, comments | `golang-naming`, `golang-code-style` | `make fmt-check` |
| Error definition, wrapping, translation | `golang-error-handling` | `make lint` |
| Goroutines, channels, locks, cancellation | `golang-concurrency`, `golang-context` | `make race` |
| Repositories, queries, migrations | `golang-database` | `make test-integration` |
| Test design and coverage | `golang-testing` | `make test-all` |
| Cobra commands and Viper configuration | `golang-spf13-cobra`, `golang-spf13-viper` | `make build` |
| Logs, metrics, traces, profiling | `golang-observability` | `make test` |
| Security-sensitive or nil-prone code | `golang-security`, `golang-safety` | `make vuln` |
| Profiling and optimization | `golang-benchmark`, `golang-performance` | `go test -bench=. ./...` |
| Debugging a defect | `golang-troubleshooting` | `make test` |
| Container engine setup, compose services, image builds, testcontainers plumbing | `podman-operations` | `make container` |
| Authoring or editing a skill | `agent-skills` | `make skills-check` |

Finish any of these with `make check`.

## Project Architecture

### Code Generation Strategy
This project uses **ogen** for OpenAPI-first API development:
- OpenAPI spec: `api/openapi.yml` - Single source of truth for API definition
- Generated code: `internal/api/v1/oas_*_gen.go` - Auto-generated from OpenAPI spec
- Generation trigger: `go generate` or `make generate`
- **Never edit generated files directly** - all API changes must be made in the OpenAPI spec

### Application Structure
- **Entry point**: `main.go` → `cmd/` - Cobra commands and process-level concerns
- **Composition root**: `internal/app/` - dependency assembly and application lifecycle
- **Domain contracts**: `internal/domain/` - business data, persistence inputs, and business-facing errors
- **Business modules**: `internal/modules/` - product capabilities; the example user repository lives in `internal/modules/user/`
- **Platform**: `internal/platform/` - configuration, database, HTTP server and middleware, observability, and rate limiting
- **Controllers**: `internal/controller/` - generated transport-to-domain mapping; not a business-logic home
- **Authentication**: `internal/auth/` - JWT validation and generated security-handler integration
- **Generated transport**: `internal/api/v1/` - ogen output; never edit directly

### Dependency Boundaries
- `internal/domain/` is an internal leaf and must not import application, module, platform, controller, or generated transport packages.
- `internal/platform/` contains technical infrastructure with no business meaning. It must not import `internal/domain/` or `internal/modules/`.
- `internal/modules/` may import domain contracts and selected platform capabilities. It must never import `internal/api/`.
- Business orchestration depends on small consumer-side interfaces. Concrete infrastructure is constructed in `internal/app/` or a CLI composition root and injected.
- Repository adapters may use the established platform database boundary, but domain data and errors exposed to consumers come from `internal/domain/`.
- Generated transport is quarantined to edge packages (`internal/controller/`, `internal/auth/`, `internal/platform/http/server/`) and the composition root.
- Do not add `utils`, `helpers`, `common`, or `shared` grab-bag packages. Name a package for the business capability or platform service it owns.
- `make architecture-check` verifies these rules and is part of `make check`.

See [docs/architecture/module-platform-boundaries.md](docs/architecture/module-platform-boundaries.md)
and the `go-project-layout` skill for the complete placement and review rules.

### Key Dependencies
- **Web framework**: Chi v5 for HTTP routing and middleware
- **Database**: PostgreSQL with pgx/v5, sqlc, squirrel, and Goose
- **API generation**: ogen for OpenAPI code generation
- **Configuration**: Viper for config management, Cobra for CLI
- **Logging**: Zap for structured logging
- **Observability**: Prometheus metrics, OpenTelemetry tracing, pprof debugging

### Development Workflow
1. **Place the change**: use `go-project-layout`; new business capabilities go under `internal/modules/`
2. **Specify business behavior**: use `backend-spec` for a new or behaviorally changed module
3. **API changes**: modify `api/openapi.yml` first
4. **Generate code**: run `make generate` to update generated API and SQL code
5. **Implement transport mapping**: update `internal/controller/` without moving business logic into it
6. **Verify boundaries and behavior**: run focused tests, then `make check`
7. **Build**: use `make build` for a local binary, `make container` for a Podman-built image, or `make release-snapshot` for release archives

### Middleware Stack
The web server includes these middleware layers (in order):
- Request timeout (60s by default), RequestID, structured panic recovery
- Request-body limit and custom logging for API routes
- Reserved-path-aware static file serving and CORS
- Trusted-proxy-aware rate limiting on the generated login handler
- Built-in routes: `/metrics` (Prometheus), `/debug/pprof/*` (profiling)

### Configuration
- CLI flags and environment variables supported via Viper
- Configuration structure defined in `internal/platform/config/config.go`
- Webserver settings: address, port, timeouts configurable via flags or env vars

## Engineering Policy

### Priority Order

When requirements compete, use this order:

1. Correctness and safety
2. Approved acceptance criteria
3. Simplicity and clarity
4. Maintainability
5. Performance, extensibility, and reuse only when required or supported by evidence

Do not add functionality, abstraction, configurability, or infrastructure merely because
it may be useful later.

### Scope and Specification

- Implement only behavior required by an approved specification, acceptance criterion,
  bug report, or explicit user request.
- Treat acceptance criteria as the implementation boundary, not a minimum from which to
  expand.
- Record non-goals for features large enough to have a specification.
- Do not add adjacent enhancements, speculative edge cases, or unrelated refactoring
  without approval.
- When requirements are materially ambiguous, resolve the ambiguity before choosing
  behavior.
- A feature is complete when its acceptance criteria are satisfied and verified. Do not
  continue adding improvements after that point.

### Simplicity

- Choose the least complex design that correctly satisfies the specified behavior and
  known operational bounds.
- Prefer explicit, readable code over clever, compressed, or highly generic code. A few
  straightforward lines are preferable to less code that is harder to understand.
- Prefer existing project patterns, the Go standard library, and current dependencies
  before introducing new abstractions or dependencies.
- Do not create an interface, framework, helper layer, factory, configuration option, or
  extension point without a current concrete need.
- Accept small, local duplication when extracting it would obscure behavior or create an
  abstraction without a stable concept.
- Minimize states, branches, conversions, callbacks, and indirection.
- Comments should explain constraints or reasons, not restate the code.

### Complexity Gates

Explicitly justify any change that introduces one or more of the following:

- A new external dependency
- A new package or architectural layer
- A new interface with only one expected implementation
- Concurrency, background work, caching, retries, or queues
- New persistent state or configuration
- A reusable abstraction created for hypothetical future consumers
- Behavior not traceable to an acceptance criterion

If the addition materially expands product scope, operational behavior, or maintenance
cost, obtain approval before implementing it.

### Bounded and Resilient Behavior

- Define and enforce relevant input, size, time, pagination, concurrency, and resource
  bounds at system boundaries.
- Reject invalid or unsupported input explicitly.
- Propagate cancellation and deadlines through blocking operations.
- Bound retries, concurrency, buffering, and memory growth; never introduce an unbounded
  mechanism.
- Preserve useful error context while exposing only the error contract required at each
  boundary.
- Clean up acquired resources reliably.
- Add resilience mechanisms only for a documented failure mode. Do not add retries,
  fallbacks, circuit breakers, or recovery paths speculatively.
- Prefer a clear failure over silently degraded or partially correct behavior unless
  degradation is explicitly specified.

## Package Contracts and Isolation

Treat every package as an independently understandable unit with an explicit contract.
This is contract-first package isolation, not a requirement to create an interface for
every concrete type.

### Package Boundaries

- A package may depend only on another package's exported contract. Do not rely on its
  internal representation, undocumented behavior, initialization order, or implementation
  details.
- Every package must have a clear purpose, ownership boundary, and set of non-goals.
- Export only what consumers need. An exported symbol is a compatibility commitment, not
  a convenience.
- Do not expose mutable internal state.
- Validate untrusted or externally derived values where they enter the trusted domain.
  Do not repeatedly validate established internal invariants without a documented reason.
- Keep implementations local to the contract package when there is one natural
  implementation. When multiple real variants exist, keep the shared contract and types
  in the parent package and place implementations in clearly named subpackages.

### Package Documentation

- Every nontrivial package must have meaningful Go package documentation describing its
  purpose and contract.
- Document relevant inputs, outputs, invariants, bounds, error behavior, concurrency
  guarantees, and resource ownership.
- Business modules and packages with substantial contracts, multiple implementations, or
  important operational behavior must include a concise human-readable `README.md`.
- A package README should explain purpose, non-goals, usage, bounds, and implementation
  structure. Do not duplicate symbol-level Go documentation.

### Interfaces and Implementations

- A documented package contract does not require a Go interface.
- Define an interface when there is a real substitution boundary, multiple
  implementations, or a consumer that needs a narrower contract.
- Keep interfaces small and behavior-oriented. Do not mirror an implementation's entire
  method set.
- Prefer consumer-defined interfaces. Shared cross-package contracts may live in a
  neutral contract or domain package.
- Constructors normally return concrete types. Consumers accept the narrow interface they
  require.
- Add compile-time interface assertions for concrete implementations.
- Use one reusable contract test suite across implementations when it protects meaningful
  shared behavior.

## Testing Policy

Tests exist to protect useful behavior, not to increase test counts or coverage
percentages.

### Test Selection

Add a test only when it protects at least one of:

- An acceptance criterion or public contract
- A meaningful boundary or validation rule
- A significant failure mode owned by this code
- A previously observed regression
- A security, concurrency, persistence, or integration invariant
- Logic whose behavior would not be obvious during review

For each proposed test, be able to state what realistic defect it would catch. If no
meaningful defect can be identified, omit the test. Use the smallest set of tests that
gives confidence in the changed behavior, and do not create separate tests that prove the
same invariant through equivalent inputs.

### Test Boundaries

- Test observable behavior and contracts, not private implementation structure.
- Prefer the narrowest stable public boundary that proves the behavior.
- Use integration tests when the important risk lies in a database, generated API
  contract, filesystem, network protocol, or external boundary.
- Use unit tests for meaningful isolated logic.
- Do not require unit, integration, and end-to-end tests for every change. Select the
  least expensive level that can detect the relevant failure.
- Prefer real values and lightweight fakes at true boundaries. Avoid deep mock graphs that
  duplicate implementation details.
- Use table-driven tests only when cases share behavior and the table is clearer than
  individual tests.
- Extract test helpers only when they reduce noise without hiding the behavior being
  verified.

### Tests to Avoid

Unless they protect a project-specific contract, do not test:

- Generated code
- Trivial getters, setters, constructors, or constant assignments
- Go standard-library or third-party-library behavior
- Private call order or internal implementation details
- Impossible states already excluded by types or validated boundaries
- Mere struct shape when the compiler already enforces it
- Duplicate happy paths that add no new risk coverage

### Test Quality

- Tests must be deterministic, isolated, and safe to run repeatedly.
- A test failure must identify the broken behavior clearly.
- Keep setup smaller and simpler than the behavior under test.
- Avoid timing assumptions, arbitrary sleeps, random inputs without fixed seeds, and
  shared mutable global state.
- Add a regression test for a bug when it can reproduce the defect and prevent recurrence
  at a stable boundary.
- Delete obsolete, redundant, or implementation-coupled tests.
- Do not use coverage percentage as a completion target. Coverage may reveal unexamined
  code, but uncovered lines alone do not justify tests.

## Change Completion

Before reporting a change complete:

- Map implemented behavior to its acceptance criteria or stated request.
- Confirm no material behavior was added beyond the approved scope.
- Run the smallest relevant verification set, expanding it when risk warrants.
- State which meaningful behaviors were tested and why.
- State any intentionally unsupported behavior or relevant tests not added.
- Run the repository-required generation, formatting, lint, and test checks appropriate
  to the changed files.

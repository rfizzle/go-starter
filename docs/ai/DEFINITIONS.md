# AI Agent Definitions and Standards

Shared terminology and decision criteria for the role agents in
[`prompts/`](prompts/). The authoritative engineering policy, testing policy,
and skill routing live in [AGENTS.md](../../AGENTS.md); this document only
names the concepts those agents refer to.

## Core Concepts

### Package Stages

**Definition**: A package stage is a complete, testable unit of work scoped to
one package, which can be validated and committed on its own.

**Completion Criteria**:
- Code compiles without errors
- The required test tiers pass (`make test-unit`, plus `make test-integration`
  or `make test-all` when the risk warrants it)
- Linting and formatting standards met
- Package documentation updated for new or materially changed contracts
- OpenAPI spec updated if API behavior changed
- Generated code is current according to `make generate-check`

**Stage Boundaries**:
- **Feature stage**: one package's share of a feature — contract, behavior, and
  tests
- **Bug fix stage**: complete resolution of one defect, including a regression
  test at a stable boundary
- **Refactor stage**: complete refactoring of a specific component or pattern

### Plan Step and Execution Step

Each role agent supports planning and execution. This is a workflow
distinction, not a model choice — which model runs which step is a client-side
decision and is not specified in this repository.

**Plan first when**:
- The change spans multiple files, packages, or components
- Architectural or placement decisions are required
- The territory is unfamiliar
- The risk is high: security, persistence, concurrency, or breaking changes
- The requester asked for an approach before edits

**Execute directly when**:
- The change is mechanical and decision-free (typo, version bump, single
  configuration value)
- An approved plan already covers the work
- The requester explicitly asked for immediate execution on a simple task

**Decision framework**:
```
Trivial, mechanical, no decisions required?
├─ Yes → execute directly
└─ No  → plan, get approval, then execute

Unsure → plan
```

### Template Validation Levels

- **Level 1 — Existence**: the template file exists at the expected path.
  Suitable for optional templates with a fallback.
- **Level 2 — Structure**: the template contains its required sections. Use for
  templates an agent's output depends on.
- **Level 3 — Completeness**: all required placeholders are present and match
  current usage. Use for templates that directly shape deliverable quality.

## Terminology

- **Skill**: a portable task guide in `.agents/skills/`, loaded on demand and
  paired with a verification command in the AGENTS.md routing table
- **Role agent**: a scoped subagent defined in `docs/ai/prompts/` and surfaced
  through `.claude/agents/`
- **Implementation Plan**: an ordered task breakdown for a feature or fix,
  returned by the planning agent for approval before development starts
- **Feature Specification**: the approved statement of behavior, acceptance
  criteria, and non-goals, kept as a co-located `<package>.spec.md` (or a
  `docs/flows/` flow spec) per the `backend-spec` skill
- **Package Review**: review of all changes within a single package stage
- **Quality Gate**: a check that must pass before work proceeds
- **Handoff Criteria**: conditions that must hold before passing work on

### Workflow States

**Planning** · **Implementing** · **Reviewing** · **Documenting** ·
**Validating**

## Quality Standards Reference

### Code Quality Gates
1. **Compilation**: code must compile without errors
2. **Testing**: the required unit and integration tiers pass (`make test-unit`,
   plus `make test-integration` or `make test-all` when the risk requires it)
3. **Repository checks**: architecture, documentation, skill-library,
   generation, formatting, vet, lint, and unit checks pass (`make check`)
4. **Race checks**: `make race` passes for concurrency, lifecycle, middleware,
   storage, or test-infrastructure changes
5. **Generation**: generated code is current (`make generate-check`)
6. **Coverage**: `make cover` clears the maintained-code floor in
   `scripts/coverage-baseline.env`

### Documentation Requirements
- Nontrivial packages carry a package doc comment describing their contract
- Exported symbols have godoc comments
- README updates for user-facing changes
- OpenAPI spec updates for API changes

### Review Criteria Priority

Mirrors the priority order in AGENTS.md:

1. **Correctness and safety**
2. **Approved acceptance criteria**
3. **Simplicity and clarity**
4. **Maintainability**
5. **Performance, extensibility, and reuse** — only when required or supported
   by evidence

## Handoff Requirements

When one agent hands work to another, include:
- a summary of completed work and the state of the working tree
- outstanding issues, blockers, or open decisions
- the acceptance criteria the next step must satisfy
- the verification commands already run and their results

## Failure Reporting

An agent that cannot meet a gate reports the failure and stops. It does not
lower the bar, skip verification, or narrow the requested scope on its own.
Report the command, its output, and what remains unfinished.

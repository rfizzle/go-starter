# AI Agent Guide

How to pick a role agent and where the boundaries between them sit.

Read [AGENTS.md](../../AGENTS.md) first. It carries the repository's
architecture, engineering policy, testing policy, and the task-type → skill →
verification routing table. Nothing in this guide overrides it.

## Skills come first

Routing starts with the skill library in [`.agents/skills/`](../../.agents/skills/README.md),
not with an agent. For any change, load the narrowest skill that covers it and
run the verification command AGENTS.md pairs with it — for example
`go-project-layout` plus `make architecture-check` for a package move, or
`golang-database` plus `make test-integration` for a repository change.

Role agents are a scoping tool on top of that. They are worth invoking when a
piece of work is long enough that keeping it inside one responsibility helps,
or when it produces a document that belongs in a specific directory. A skill
still applies inside whichever agent is running.

## The four role agents

| Agent | Focus | Produces |
| --- | --- | --- |
| [Documentation](prompts/documentation.md) | Specs, READMEs, prompts, templates | Documents in `docs/features/` and `docs/ai/` |
| [Planning](prompts/planning.md) | Task breakdown, risks, sequencing | Plans in `docs/plans/` |
| [Review](prompts/review.md) | Code quality, security, standards | Reviews in `docs/reviews/` |
| [Development](prompts/development.md) | Implementation and fixes | Source changes |

These definitions live in `docs/ai/prompts/` and are surfaced to Claude Code as
relative symlinks under `.claude/agents/`. Regenerate the links with
`make agents-setup`; `make skills-check` fails on a missing or dangling one.

### Documentation agent

- **Owns**: feature specifications, README files, prompts and templates under
  `docs/ai/`, and user-facing guides.
- **Does not**: write application code, create implementation plans, or perform
  code review.
- **Hands off to**: Planning when a spec is ready for breakdown; Development
  when documentation needs verified code examples.

### Planning agent

- **Owns**: turning a feature document into an ordered, package-scoped plan
  with dependencies, risks, and a verification strategy.
- **Does not**: write application code or gather requirements.
- **Hands off to**: Development when the plan is complete; Documentation when
  requirements are too thin to plan against.

### Review agent

- **Owns**: analysis of code changes against the engineering and testing policy
  in AGENTS.md, recorded as a structured document in `docs/reviews/`.
- **Does not**: modify application code or write plans.
- **Hands off to**: Development for the fixes it identifies; Planning when a
  finding implies an architectural change.

### Development agent

- **Owns**: implementation, bug fixes, refactors, and the tests that protect
  them.
- **Does not**: author specifications or write its own review documents.
- **Hands off to**: Review when an implementation is ready for assessment.

## Task type mapping

| Task | Agent | Start with skill |
| --- | --- | --- |
| Feature specification, user stories | Documentation | `backend-spec` |
| README or guide updates | Documentation | — |
| Implementation plan, task breakdown | Planning | `go-project-layout` |
| Code quality or standards review | Review | the skill matching the changed area |
| Security review | Review | `golang-security` |
| Feature implementation | Development | the AGENTS.md routing table |
| Bug fix | Development | `golang-troubleshooting` |
| Package placement or move | Development | `go-project-layout`, `project-conventions` |
| Authoring or editing a skill | Documentation | `agent-skills` |

## Plan then execute

Plan before editing for anything non-trivial: multiple files or packages,
architectural decisions, unfamiliar territory, high-risk changes such as
security or persistence, or any multi-step process. Present the approach, get
it approved, then execute it.

Go straight to execution for mechanical changes — a typo, a version bump, a
single-line configuration tweak — or when following a plan that was already
approved. When the complexity is unclear, plan.

Which model runs which step is a client-side decision and is not specified in
this repository.

## Handoffs

The natural chain is Documentation → Planning → Development → Review, with
Review looping back to Development for fixes. A handoff should carry:

- what was completed and what state the working tree is in;
- outstanding issues, blockers, or open decisions;
- the acceptance criteria the next step has to satisfy;
- the verification commands already run and their results.

Before handing off, confirm the current scope is actually finished, the
relevant verification from AGENTS.md has been run, and any behavior added
beyond the approved scope has been removed.

## Quality gates

Every agent's work is subject to the same gates, defined in
[DEFINITIONS.md](DEFINITIONS.md) and enforced by `make check`. An agent that
cannot meet them reports the failure and stops rather than lowering the bar.

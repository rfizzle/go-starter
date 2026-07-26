# Workflow Shortcuts

Common multi-step patterns for AI-assisted work. Each step still loads the
narrowest applicable skill and runs the verification command that
[AGENTS.md](../../AGENTS.md) pairs with it.

## Express patterns

Single agent, straight to execution:

| Task | Agent | Skill | Verify |
| --- | --- | --- | --- |
| Simple bug fix | Development | `golang-troubleshooting` | `make test` |
| Documentation update | Documentation | — | `make docs-check` |
| Quick code review | Review | area-specific | — |
| Naming or formatting cleanup | Development | `golang-naming`, `golang-code-style` | `make fmt-check` |

## Guided patterns

Plan first, then execute after approval.

### New feature

```
1. Documentation → feature spec in docs/features/     (skill: backend-spec)
2. Planning      → package-scoped plan in docs/plans/ (skill: go-project-layout)
3. Development   → design the approach, get approval
4. Development   → implement                          (skill: per AGENTS.md routing)
5. Review        → validate against the plan          (verify: make check)
```

### Major refactor

```
1. Review      → analyze the current state
2. Planning    → sequencing, risks, rollback
3. Development → design the new structure   (skills: go-project-layout, project-conventions)
4. Development → execute in package-sized steps
5. Review      → confirm boundaries hold    (verify: make architecture-check, make check)
```

### Complex defect

```
1. Review      → reproduce and locate       (skill: golang-troubleshooting)
2. Development → design the fix and a regression test at a stable boundary
3. Development → implement                  (verify: make test, make race if concurrent)
4. Review      → confirm the regression test fails without the fix
```

### API change

```
1. Edit api/openapi.yml — the specification is the source of truth
2. make generate                            (never edit internal/api/v1 by hand)
3. Development → map transport to domain in internal/controller/
4. Verify: make generate-check, make check
```

### Database change

```
1. Planning    → migration strategy and backfill, if any
2. Development → migration, queries, repository (skill: golang-database)
3. make generate — regenerate sqlc output
4. Verify: make test-integration
```

## Choosing an entry point

```
Is the work already specified?
├─ No  → Documentation agent, plan mode
└─ Yes → Is the sequencing clear?
         ├─ No  → Planning agent
         └─ Yes → Development agent
```

```
Is the change mechanical, single-file, and decision-free?
├─ Yes → execute directly
└─ No  → plan first, get approval, then execute
```

## Referencing prior work

Point an agent at existing artifacts instead of re-deriving them:

```
"Follow the plan in docs/plans/<name>.md"
"Apply the findings from docs/reviews/<name>.md"
"Match the pattern in internal/modules/user/"
```

## When something goes wrong

| Symptom | Response |
| --- | --- |
| Agent refuses the task as out of scope | Check [agent-guide.md](agent-guide.md) and switch to the agent that owns it |
| Plan keeps growing | Cut it to the acceptance criteria; record the rest as non-goals |
| Verification fails | Stop, report the output, do not commit — see AGENTS.md, Change Completion |
| Change drifts beyond the request | Revert the excess; scope reduction is the requester's call |
| Generated file needs a change | Change its source (`api/openapi.yml`, `queries/`) and regenerate |

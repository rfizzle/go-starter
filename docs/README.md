# Go Starter · documents

Each file answers one question. Facts live here once; `AGENTS.md` says where
the code is; skills say how a change is made.

| File | Question it answers | Changes when |
|---|---|---|
| `development/ci-contract.md` | What does each CI workflow enforce, and how do I reproduce a failure? | a workflow or a tool pin changes |
| `development/deployment.md` | How are releases and images built and deployed? | the release or deployment process changes |
| `architecture.md` | What are the big shapes, and why those? | a structural shape changes |
| `decisions.md` | What was decided, when and why? | a decision is made or superseded |
| `research/` | Dated snapshots | never; a new look is a new dated file |

- Documents name no code symbol; code cites documents by path and heading.
  `make docs-check` fails on a citation that does not resolve and lists
  headings nothing cites.
- Anything time-bound goes in `research/` with the date in the filename and
  is not edited afterwards.
- Planning is not in this repository. Nothing here names a story, ticket or
  sprint.

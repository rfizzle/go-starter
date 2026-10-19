# Go Starter · documents

Each file answers one question. Facts live here once; `AGENTS.md` says where
the code is; skills say how a change is made.

| File | Question it answers | Changes when |
|---|---|---|
| `development/README.md` | Which guide covers the boundary I am changing? | a guide is added or removed |
| `development/environment.md` | How do I set up and run the service locally? | a prerequisite or local service changes |
| `development/database.md` | How is PostgreSQL code written, migrated and tested here? | the database tooling or patterns change |
| `development/ogen-workflow.md` | How does an API change flow from the OpenAPI spec to a controller? | the generation workflow changes |
| `development/ci-contract.md` | What does each CI workflow enforce, and how do I reproduce a failure? | a workflow or a tool pin changes |
| `development/deployment.md` | How are releases and images built and deployed? | the release or deployment process changes |
| `development/project-context.md` | What does the service do today, and with what? | the implemented surface changes |
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

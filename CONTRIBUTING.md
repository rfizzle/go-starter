# Contributing

Thanks for improving Go Starter. Keep changes focused, independently testable,
and consistent with the OpenAPI-first workflow.

## Before you start

Participation is governed by [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

Read [AGENTS.md](AGENTS.md) for repository rules and the
[README quickstart](README.md#five-minute-quickstart) for prerequisites. Run
`make tools` to install the pinned development toolchain, `make env` to verify
its versions, and `make help` to inspect the supported workflows.

Create a topic branch and avoid mixing generated output, dependency updates,
format-only edits, and behavioral changes unless they are part of the same
change.

## Development workflow

The commands, and which to run for which change, are in the Commands section
of [AGENTS.md](AGENTS.md#commands); finish with `make check`.

`make ci` runs the same read-only gates the pipeline enforces. The
[CI contract](docs/development/ci-contract.md) records what each workflow
checks, how release provenance is produced, and how to reproduce a failure
locally.

Never edit ogen files under `internal/api/v1/` or sqlc files under
`internal/platform/database/postgres/db/` directly. The
[api-contract](.agents/skills/api-contract/SKILL.md) and
[golang-database](.agents/skills/golang-database/SKILL.md) skills, and the
invariants in [AGENTS.md](AGENTS.md#invariants), provide more detail.

Use a conventional commit subject such as `feat(auth): add token rotation` or
`fix(storage): preserve transaction errors`. Keep pull requests small, explain
the user-visible outcome, identify verification performed, and call out
migrations, compatibility concerns, or follow-up work. The pull request
template in `.github/PULL_REQUEST_TEMPLATE.md` lists the expected checks; add
an entry to [CHANGELOG.md](CHANGELOG.md) when a change requires action from
someone using the template.

Do not include credentials, private vulnerability details, generated build
artifacts, or unrelated worktree changes. Follow [SECURITY.md](SECURITY.md) for
security reports.

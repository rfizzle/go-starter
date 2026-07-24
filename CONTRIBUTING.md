# Contributing

Thanks for improving Go Starter. Keep changes focused, independently testable,
and consistent with the OpenAPI-first workflow.

## Before you start

Read [AGENTS.md](AGENTS.md) for repository rules and the
[README quickstart](README.md#five-minute-quickstart) for prerequisites. Run
`make tools` to install the pinned development toolchain, `make env` to verify
its versions, and `make help` to inspect the supported workflows.

Create a topic branch and avoid mixing generated output, dependency updates,
format-only edits, and behavioral changes unless they are part of the same
change.

## Development workflow

1. Add or update tests for behavioral changes.
2. Change `api/openapi.yml` before changing an API contract.
3. Change SQL query sources or add a new migration before regenerating sqlc.
4. Run `make generate` after generator inputs change.
5. Run `make fmt` for changed Go or YAML files.
6. Run `make architecture-check`, `make docs-check`, and focused package
   tests, then `make check`.
7. Run `make race` for concurrency, lifecycle, middleware, storage, or test
   infrastructure changes.
8. Confirm `git diff --check` is clean and review the staged diff.

Never edit ogen files under `internal/api/v1/` or sqlc files under
`internal/platform/database/postgres/db/` directly. The
[ogen workflow](docs/development/ogen-workflow.md),
[database guide](docs/development/database.md), and
[Go standards](docs/guidelines/go-standards.md) provide more detail.

Use a conventional commit subject such as `feat(auth): add token rotation` or
`fix(storage): preserve transaction errors`. Keep pull requests small, explain
the user-visible outcome, identify verification performed, and call out
migrations, compatibility concerns, or follow-up work.

Do not include credentials, private vulnerability details, generated build
artifacts, or unrelated worktree changes. Follow [SECURITY.md](SECURITY.md) for
security reports.

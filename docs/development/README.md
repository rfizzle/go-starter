# Development Documentation

This directory contains the maintained workflows for developing Go Starter.
Start with the root [quickstart](../../README.md#five-minute-quickstart), then
use the focused guide for the boundary you are changing.

## Reading order

1. [project-context.md](project-context.md) — implemented architecture and
   workflow
2. [environment.md](environment.md) — pinned tools, Compose, configuration,
   migrations, and seeding
3. [ogen-workflow.md](ogen-workflow.md) — OpenAPI and generated transport
4. [database.md](database.md) — PostgreSQL, sqlc, squirrel, and migrations
5. [ci-contract.md](ci-contract.md) — what each workflow enforces and how to
   reproduce it locally

The accepted package-placement rules live in
[module-platform-boundaries.md](../architecture/module-platform-boundaries.md).
Engineering, package-contract, and testing policy live in
[AGENTS.md](../../AGENTS.md) and the `.agents/skills/` library.

## Supported commands

```sh
make help               # List the complete supported target set
make generate           # Regenerate ogen and sqlc output
make architecture-check # Verify package placement and import direction
make docs-check         # Verify documentation links, commands, paths, and version
make fmt                # Format Go and YAML sources
make lint               # Run go vet and golangci-lint
make test-unit          # Run tests without external services
make test-integration   # Run tests with disposable service containers
make test-all           # Run both test tiers
make race               # Run the Go suite with the race detector
make cover              # Report and gate maintained-code coverage
make check              # Run the fast read-only repository gates
make secrets-check      # Scan the full git history for committed secrets
make ci                 # Run the gates CI enforces (check + secrets-check + vuln)
make build              # Build bin/gostarter for the current platform
```

`make generate`, `make fmt`, and `make tidy` intentionally modify the
worktree. `make check` is read-only when generated source and
formatting are current.

## Standard change workflow

1. Bound the requested behavior and non-goals.
2. Use the placement table in the architecture decision.
3. Change `api/openapi.yml` first for HTTP contract changes.
4. Change migrations or query sources first for database changes.
5. Run `make generate` after changing generator inputs.
6. Implement the smallest change that satisfies the contract.
7. Run focused tests, then `make check`; add `make race` or
   `make test-integration` when the affected risk warrants it.
8. Review `git diff --check` and the staged diff before committing.

Never edit files under `internal/api/v1/` or
`internal/platform/database/postgres/db/` directly.

## Placement summary

| Change | Location |
| --- | --- |
| HTTP contract | `api/openapi.yml` |
| Generated ogen transport | `internal/api/v1/` |
| Transport/domain mapping | `internal/controller/` |
| Shared business data or contracts | `internal/domain/` |
| Business behavior or repository adapter | `internal/modules/<capability>/` |
| PostgreSQL lifecycle, migration, or generated SQL | `internal/platform/database/` |
| HTTP server or middleware | `internal/platform/http/` |
| Configuration or telemetry | matching `internal/platform/` package |
| Dependency construction and lifecycle | `internal/app/` or a CLI composition root |

Generated transport stays at the edge. Platform packages must not import
business modules or domain contracts. Controllers depend on small
consumer-side interfaces and do not own business rules.

## Common tasks

### Add or change an API operation

1. Edit `api/openapi.yml`.
2. Run `make generate`.
3. Implement the generated handler or security contract in
   `internal/controller/` or `internal/auth/`.
4. Add the narrowest controller or HTTP test that proves the changed contract.
5. Run `make generate-check` and `make check`.

### Change PostgreSQL behavior

1. Create a sequential migration with
   `make migration name=add_widgets` when the schema changes.
2. Edit fixed SQL under `internal/platform/database/postgres/queries/`.
3. Run `make generate`.
4. Keep pgx/sqlc mapping in the owning module repository edge.
5. Use the disposable PostgreSQL harness for schema, query, constraint, lock,
   or transaction behavior.

### Validate a documentation-only change

Run:

```sh
make docs-check
make check
git diff --check
```

The documentation check validates local Markdown links, documented Make
targets, required workflow paths, retired path names, and the Go version
declared in `go.mod`.

## Local service troubleshooting

Use [environment.md](environment.md) for the supported Compose flow. The
shortest diagnostics are:

```sh
podman compose config
podman compose ps
podman compose logs postgres
```

Use `make dev-reset` only when intentionally deleting the named local
PostgreSQL volume.

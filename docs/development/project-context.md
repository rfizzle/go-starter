# Project Context

Go Starter is an OpenAPI-first foundation for PostgreSQL-backed HTTP services.
It demonstrates a complete but deliberately small vertical slice: explicit
configuration, migration and seed commands, credential login, JWT bearer
authentication, a protected current-user operation, health probes, metrics,
optional OpenTelemetry export, and graceful shutdown.

It is a starter rather than a finished production deployment. Downstream
projects must choose their own secrets, CORS, TLS termination, database sizing,
monitoring, backup, and availability policies.

## Sources of truth

| Concern | Source |
| --- | --- |
| Go toolchain | `go.mod` |
| HTTP contract | `api/openapi.yml` |
| Generated HTTP code | `internal/api/v1/` via `make generate` |
| SQL schema | `internal/platform/database/postgres/migrations/` |
| Fixed SQL queries | `internal/platform/database/postgres/queries/` |
| Generated SQL code | `internal/platform/database/postgres/db/` via `make generate` |
| Tool versions | `scripts/tool-versions.env` |
| Release definition | `.goreleaser.yaml` |
| Runtime configuration | `internal/platform/config/` and command flags |

Never edit generated ogen or sqlc files directly.

## Runtime architecture

```text
main.go
  └── cmd (Cobra + Viper composition roots)
        ├── migrate ──→ platform/database/postgres
        ├── seed ─────→ cli/seed ──→ modules/user
        └── serve ────→ internal/app
                          ├── modules/credential + modules/user
                          ├── auth + controller + api/v1
                          └── platform/{config,database,health,http,
                                      observability,ratelimiter}
```

`internal/app` constructs concrete dependencies and owns runtime shutdown
ordering. `internal/controller` maps generated transport types to small
business-facing contracts. `internal/domain` is an infrastructure-free leaf.
Business capabilities live under `internal/modules`; technical services live
under `internal/platform`.

See the accepted
[module/platform boundary decision](../architecture/module-platform-boundaries.md)
and the [current architecture map](../architecture/current-state.md).

## Implemented HTTP surface

The OpenAPI document currently defines:

- process-oriented liveness;
- dependency-aware readiness;
- credential login returning a JWT bearer token; and
- a protected current-user operation.

The server also exposes `/metrics`. Profiling under `/debug/pprof/*` is
disabled by default and is not part of the OpenAPI surface. The framework-neutral
asset under `web/build/` is served only outside reserved API, metrics, and
profiling paths.

Login attempts are limited by a bounded process-local limiter. Forwarding
headers affect client identity only when the immediate peer belongs to a
configured trusted-proxy CIDR.

## Persistence and lifecycle

The application owns one pgx pool. The user repository uses sqlc for fixed
queries and squirrel for runtime-shaped filters, mapping driver values to
`domain.User` at the module boundary.

Schema changes are embedded Goose migrations. Production deployments should
run `migrate up` as a one-shot job before replicas start; the command uses a
PostgreSQL advisory lock to serialize concurrent attempts. Startup migration is
available only as an explicit configuration policy and is disabled by default.

Readiness combines lifecycle state with a bounded PostgreSQL ping. Liveness
does not call external dependencies. Cancellation propagates through runtime
components, and shutdown drains the HTTP server before closing the database
pool.

## Configuration and observability

Cobra flags, `GOSTARTER_*` environment variables, and an explicitly selected
configuration file feed command-scoped Viper instances. Precedence and local
examples are documented in [environment.md](environment.md). Validation occurs
before network listeners or database work begins.

Zap provides structured logging. Prometheus serves HTTP, limiter, and pool
metrics. OpenTelemetry trace and metric export is optional and inactive unless
explicitly enabled with an OTLP collector.

## Development workflow

1. Read `AGENTS.md` and the applicable repository skill.
2. Place the change according to the architecture decision.
3. Update the OpenAPI or SQL source before generated output.
4. Select tests that protect acceptance criteria, boundaries, regressions, or
   named operational risks.
5. Implement only the approved behavior.
6. Run focused checks, then `make check`.
7. Run `make race`, `make test-integration`, or `make test-all` when the
   affected risk requires them.

Key commands:

| Command | Meaning |
| --- | --- |
| `make build` | Build the current-platform binary in `bin/`. |
| `make generate-check` | Regenerate and fail on source drift. |
| `make architecture-check` | Enforce package and import boundaries. |
| `make docs-check` | Validate documentation references and version. |
| `make test-unit` | Run tests without external services. |
| `make test-integration` | Run disposable-service integration tests. |
| `make test-all` | Run both test tiers. |
| `make race` | Run Go tests with the race detector. |
| `make cover` | Report and gate maintained-code coverage. |
| `make check` | Run the fast repository validation suite. |
| `make release-snapshot` | Exercise release archives, SBOMs, and images without publishing. |

## Technology

- Go 1.26.5
- Chi v5 and ogen
- PostgreSQL, pgx/v5, sqlc, squirrel, and Goose
- Cobra and Viper
- Zap, Prometheus, and OpenTelemetry
- Testify and Testcontainers-Go
- GoReleaser and distroless container images

The root [README](../../README.md) is the user entrypoint. This document is the
maintained engineering summary; deeper behavior belongs in the focused guides
linked above.

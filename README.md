# Go Starter

Go Starter is an OpenAPI-first foundation for PostgreSQL-backed HTTP services.
It provides a small example domain and the operational plumbing that most Go
services need, while keeping the API contract, generated transport code, domain
logic, and storage implementation separate.

The module currently requires Go 1.26.5. Its main stack is:

- Cobra and Viper for commands and layered configuration
- Chi and ogen for HTTP routing and generated OpenAPI types
- PostgreSQL, pgx/v5, sqlc, and Goose for persistence
- JWT bearer authentication
- Zap logging, Prometheus metrics, and optional OpenTelemetry export
- testify and pgxmock for tests

## What is included

The production-oriented foundations are implemented and wired into `serve`:
early configuration validation, structured logging, graceful shutdown,
PostgreSQL pooling and migrations, JWT validation, health endpoints, Prometheus
metrics, request limits, configurable TLS, and optional OTLP export. This is a
starter, not a production deployment you should ship unchanged: choose
deployment-specific secrets, CORS origins, connection limits, TLS termination,
monitoring, backup, and availability policies first.

The user domain, the `admin@example.com` seed record, and embedded static files
are examples intended to be replaced or extended. The rate-limiter package is
also an optional example and is not installed in the active middleware stack.
OpenTelemetry export is inactive unless explicitly enabled and given access to
an OTLP collector.

## Five-minute quickstart

You need Git, Go 1.26.5, Make, and Docker (or another way to run PostgreSQL).

Clone the repository, verify the toolchain, build the binary, and inspect the
available commands:

```sh
git clone https://github.com/rfizzle/go-starter.git
cd go-starter
go version
make tools
make env
make build
./bin/gostarter --help
```

Create the ignored local configuration file, then start the complete
development flow:

```sh
cp .env.example .env
make dev-run
```

`dev-run` starts the pinned PostgreSQL Compose service, waits for its health
check, loads `.env`, applies pending embedded migrations, idempotently seeds the
example user, and serves the application. In another terminal, probe it:

```sh
curl --fail http://localhost:8080/api/v1/healthz/liveness
curl --fail http://localhost:8080/api/v1/healthz/readiness
curl --fail http://localhost:8080/metrics
```

Stop the service with Ctrl-C, then stop PostgreSQL while preserving its named
data volume:

```sh
make dev-down
```

Run `make dev-up` when only PostgreSQL is needed. `make dev-reset` is the
explicit destructive path: it deletes the local PostgreSQL volume and starts a
fresh healthy database. The credentials in `compose.yaml` and `.env.example`
are for development only; do not reuse them in a deployed environment.

## Configuration

Each command accepts `--config PATH`, but no default configuration file or
search path is configured. For values used by `serve` and `seed`, precedence is:

1. command-line flags;
2. `GOSTARTER_*` environment variables;
3. the file passed with `--config`;
4. command defaults.

Nested keys replace dots and dashes with underscores. For example,
`database.uri` maps to `GOSTARTER_DATABASE_URI`, and
`webserver.read-header-timeout` maps to
`GOSTARTER_WEBSERVER_READ_HEADER_TIMEOUT`. `--verbose` is a shortcut that
forces debug logging. Run `./bin/gostarter serve --help` or
`./bin/gostarter seed --help` for the complete command-specific settings.

Configuration is validated before network or database resources are opened.
In particular, `serve` requires a PostgreSQL URI and a JWT secret of at least
32 bytes. Secrets are never included in validation errors.

## Core commands

| Command | Purpose |
| --- | --- |
| `make help` | List supported targets. |
| `make tools` | Install the repository-pinned development tools. |
| `make env` | Report installed and required tool versions. |
| `make build` | Build `bin/gostarter` without changing source files. |
| `make check` | Run generation drift, formatting, vet, lint, and test checks. |
| `make test` | Run all Go tests. |
| `make race` | Run all Go tests with the race detector. |
| `make cover` | Write `coverage.out` and print statement coverage. |
| `make generate` | Regenerate ogen and sqlc output. |
| `make generate-check` | Fail if committed generated code is stale. |
| `make dev-up` | Start local PostgreSQL and wait for health. |
| `make dev-run` | Start PostgreSQL, migrate, seed, and serve using `.env`. |
| `make dev-down` | Stop local services without deleting database data. |
| `make dev-reset` | Delete local database data and start PostgreSQL fresh. |
| `make release-snapshot` | Build local release archives in `dist/`. |
| `make docker` | Build the current container image. |
| `make vuln` | Scan Go dependencies with `govulncheck`. |

`make tools` reads exact versions from `scripts/tool-versions.env`. Set
`TOOL_BIN` to install into an isolated directory; otherwise tools are installed
to `GOBIN` or the first `GOPATH` bin directory. It is the shared entrypoint for
local and CI setup, and `make env` compares the binaries on `PATH` with the same
declarations.
Mutating maintenance commands such as `make fmt`, `make generate`, `make tidy`,
and `make vendor` are intentionally explicit.

## Architecture map

| Path | Responsibility |
| --- | --- |
| `main.go`, `cmd/` | Process entrypoint and Cobra commands. |
| `pkg/config/` | Application settings and command-scoped validation. |
| `api/openapi.yml` | Source of truth for the HTTP API. |
| `internal/api/v1/` | ogen-generated transport code; never edit directly. |
| `internal/controller/` | Hand-written implementations of generated handlers. |
| `internal/app/` | Dependency assembly and application lifecycle. |
| `internal/web/` | Chi server, middleware ordering, metrics, and profiling routes. |
| `internal/auth/`, `internal/user/` | Authentication and the example user domain. |
| `internal/storage/postgres/` | pgx storage, sqlc queries, and embedded migrations. |
| `internal/observability/` | Logging and optional OpenTelemetry providers. |
| `web/` | Embedded static assets. |

See the [documentation index](docs/README.md), the
[current architecture notes](docs/architecture/current-state.md), and the
[development foundations](docs/development/foundations.md) for deeper context.

## API and generation workflow

API changes start in `api/openapi.yml`. Run `make generate`, then implement any
changed generated interfaces in `internal/controller/`. SQL changes start in
`internal/storage/postgres/queries/` or a new sequential migration, followed by
`make generate`. Never hand-edit `internal/api/v1/oas_*_gen.go` or the
sqlc-generated files in `internal/storage/postgres/db/`.

Before committing generator inputs, run:

```sh
make generate
make generate-check
make check
```

The [ogen workflow guide](docs/development/ogen-workflow.md) and
[database guide](docs/development/database.md) cover the longer workflows.

## Testing tiers

- Focused unit tests: `go test ./internal/auth` (substitute the changed package).
- Full unit/package suite: `make test`.
- Static and generated-code gates: `make check`.
- Concurrency-sensitive changes: `make race`.
- Coverage report: `make cover`.
- Dependency vulnerability scan: `make vuln`.

The [testing strategy](docs/guidelines/testing-strategy.md) describes project
conventions. Database behavior is primarily isolated behind mocks and test
doubles; use a disposable PostgreSQL instance for end-to-end validation.

## Deployment notes

`make release-snapshot` uses `.goreleaser.yaml` to produce cross-platform
archives. `make docker` builds the multi-stage image from
`build/docker/Dockerfile`; the final image runs as a non-root user and invokes
`serve`.

The service applies pending embedded migrations before accepting traffic.
Provide configuration through deployment-managed environment variables or an
explicit config file, and supply PostgreSQL separately. Plain HTTP is the
default because TLS is normally terminated by a trusted ingress or load
balancer; set both TLS file options only when the process terminates TLS itself.
Profiling is disabled by default and should remain private when enabled.

Release publication, a production Compose stack, and deployment manifests are
not included. Validate the image, health probes, secret delivery, CORS policy,
database migration concurrency, observability, and rollback strategy in your
target platform.

## Contributing and security

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow. Report
suspected vulnerabilities according to [SECURITY.md](SECURITY.md), without
placing sensitive details in a public issue.

Go Starter is available under the [MIT License](LICENSE).

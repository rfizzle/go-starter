# Development Environment

This guide covers the supported local service workflow. See the root
[README](../../README.md) for the project overview and
[CONTRIBUTING](../../CONTRIBUTING.md) for the full validation workflow.

## Prerequisites

- Go 1.26.5 (the version declared in `go.mod`)
- Make
- Podman 4.7+ with a Compose provider (`podman compose` delegates to
  `podman-compose` or a Compose v2 `docker-compose` binary on `PATH`)

Install and validate the additional tools used by generation, formatting,
linting, release, migrations, and vulnerability and secret scanning:

```sh
make tools
make env
```

The bootstrap and validation targets both read exact module versions from
`scripts/tool-versions.env`; `make tools` is also the supported CI setup
entrypoint. The bootstrap uses `GOBIN` or the first `GOPATH` bin directory by
default. To keep tools isolated, choose a directory and put it first on `PATH`:

```sh
TOOL_BIN=/absolute/path/to/project-tools make tools
PATH="/absolute/path/to/project-tools:${PATH}" make env
```

ogen remains pinned as a module dependency in `go.mod` and is run through
`go generate`, so it is not installed as a separate binary.

## Container engine

Podman is the repository's container engine: it runs the local PostgreSQL
Compose service, builds images (`make container`), and backs the
testcontainers-based integration tier. On Linux, enable the rootless API
socket once so testcontainers can reach the engine, and export its address:

```sh
systemctl --user enable --now podman.socket
export DOCKER_HOST="unix://$(podman info --format '{{.Host.RemoteSocket.Path}}')"
export TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED=true
```

On macOS, create and start the Podman machine once with `podman machine init`
and `podman machine start`; it serves the API socket automatically. The
`podman-operations` skill in `.agents/skills/` documents the complete
workflow, including multi-architecture builds and the GoReleaser boundary.

Dependabot does not watch `scripts/tool-versions.env` — none of its configured
ecosystems can see a `go install` target declared in a shell environment file.
Review those pins periodically; the [CI contract](ci-contract.md) describes the
review.

## First run

Copy the example application environment, then run the development target:

```sh
cp .env.example .env
make dev-run
```

The target performs the local startup sequence:

1. starts the `postgres` service from `compose.yaml`;
2. waits for PostgreSQL's `pg_isready` health check;
3. exports the values from the ignored `.env` file;
4. runs `go run . migrate up`, which applies pending embedded migrations under
   a PostgreSQL advisory lock;
5. runs `go run . seed`, which idempotently inserts the explicitly configured
   development user and prints a generated password once if no password was
   configured and the account is created;
6. runs `go run . serve`.

The service is ready when both health probes succeed:

```sh
curl --fail http://localhost:8080/api/v1/healthz/liveness
curl --fail http://localhost:8080/api/v1/healthz/readiness
```

Liveness reports only whether the process can answer HTTP. Readiness also pings
PostgreSQL within `database.ping-timeout`; it returns `503` during startup,
shutdown, or a database outage and recovers after PostgreSQL can serve again.
The response does not include connection details or raw database errors.

Stop the Go process with Ctrl-C. PostgreSQL continues running until you stop it:

```sh
make dev-down
```

`dev-down` preserves the named database volume, so the next `dev-up` or
`dev-run` retains local data.

## Local configuration

The committed `.env.example` uses the exact Viper environment names consumed by
the CLI. Nested configuration keys replace dots and dashes with underscores and
use the `GOSTARTER_` prefix:

| Application key | Environment variable |
| --- | --- |
| `environment` | `GOSTARTER_ENVIRONMENT` |
| `database.uri` | `GOSTARTER_DATABASE_URI` |
| `database.auto-migrate` | `GOSTARTER_DATABASE_AUTO_MIGRATE` |
| `auth.jwt-secret` | `GOSTARTER_AUTH_JWT_SECRET` |
| `rate-limit.max-requests` | `GOSTARTER_RATE_LIMIT_MAX_REQUESTS` |
| `rate-limit.window` | `GOSTARTER_RATE_LIMIT_WINDOW` |
| `rate-limit.max-keys` | `GOSTARTER_RATE_LIMIT_MAX_KEYS` |
| `rate-limit.cleanup-interval` | `GOSTARTER_RATE_LIMIT_CLEANUP_INTERVAL` |
| `rate-limit.trusted-proxies` | `GOSTARTER_RATE_LIMIT_TRUSTED_PROXIES` |
| `logging.level` | `GOSTARTER_LOGGING_LEVEL` |
| `logging.format` | `GOSTARTER_LOGGING_FORMAT` |
| `observability.enabled` | `GOSTARTER_OBSERVABILITY_ENABLED` |
| `observability.endpoint` | `GOSTARTER_OBSERVABILITY_ENDPOINT` |
| `observability.insecure` | `GOSTARTER_OBSERVABILITY_INSECURE` |
| `observability.sample-ratio` | `GOSTARTER_OBSERVABILITY_SAMPLE_RATIO` |
| `webserver.address` | `GOSTARTER_WEBSERVER_ADDRESS` |
| `webserver.port` | `GOSTARTER_WEBSERVER_PORT` |

The `.env` file is loaded by `make dev-run`; the application does not
automatically read dotenv files when invoked directly. To run commands
manually, export the file first:

```sh
set -a
. ./.env
set +a
go run . migrate up
go run . seed
go run . serve
```

Command flags take precedence over these environment variables. See
`go run . migrate --help`, `go run . serve --help`, and
`go run . seed --help` for every supported setting.
The `cmd` package owns Viper file, environment, and flag loading;
`internal/platform/config` contains only decoded values and validation.
Logger and OpenTelemetry provider construction lives in
`internal/platform/observability`. OpenTelemetry export is disabled unless
`observability.enabled` is true.

Login attempts are limited by client IP. With no trusted proxy CIDRs (the safe
default), `X-Forwarded-For` is ignored and the network peer is the limiter key.
When traffic arrives through reverse proxies, set
`GOSTARTER_RATE_LIMIT_TRUSTED_PROXIES` to a comma-separated list of canonical
CIDRs for proxies operated by the deployment. The middleware walks the
forwarding chain from the trusted peer toward the client and stops at the first
untrusted address. The limiter is in-memory and process-local, so multiple
replicas enforce independent limits rather than one global aggregate limit.

## PostgreSQL lifecycle

The Compose service uses the pinned `postgres:17.5-alpine` image, binds only to
`127.0.0.1:5432`, and stores data in a named volume. Its `gostarter` username,
password, and database are intentionally non-production values that match the
URI in `.env.example`.

| Target | Behavior |
| --- | --- |
| `make dev-up` | Start PostgreSQL and wait until it is healthy. |
| `make dev-down` | Stop Compose services and preserve database data. |
| `make dev-run` | Run `dev-up`, migrate, seed, and serve. |
| `make dev-reset` | Delete the named PostgreSQL volume, then start fresh. |

`dev-reset` is the only local development target that deletes database data.
Run it explicitly when a clean database is required:

```sh
make dev-reset
```

`migrate up` is the explicit, idempotent schema step and serializes concurrent
attempts with a PostgreSQL session advisory lock. `serve` leaves
`database.auto-migrate` disabled by default; enable it only for deployments
that deliberately choose startup migration. `seed` expects the schema to be
current, is refused outside the development environment, and requires
`GOSTARTER_SEED_EMAIL`. Set `GOSTARTER_SEED_PASSWORD` explicitly or leave it
unset to generate a password that is displayed only when the account is first
created. Existing accounts are skipped without changing or disclosing their
credentials.
Migration SQL lives in
`internal/platform/database/postgres/migrations/` and is embedded in the
binary.

## Troubleshooting

Inspect container and health-check state:

```sh
podman compose ps
podman compose logs postgres
```

If port 5432 is already in use, stop the conflicting local PostgreSQL instance
or change both the Compose port mapping and `GOSTARTER_DATABASE_URI`.

To validate the Compose model without starting containers:

```sh
podman compose config
```

If `podman compose` reports that no compose provider is available, install
`podman-compose` or a Compose v2 `docker-compose` binary.

For generated code, tests, linting, and release commands, use the targets listed
by `make help`.

# Development Environment

This guide covers the supported local service workflow. See the root
[README](../../README.md) for the project overview and
[CONTRIBUTING](../../CONTRIBUTING.md) for the full validation workflow.

## Prerequisites

- Go 1.26.5 (the version declared in `go.mod`)
- Make
- Docker with the Compose plugin

Install and validate the additional tools used by generation, formatting,
linting, release, migrations, and vulnerability scanning:

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
4. runs `go run . seed`, which applies pending embedded migrations and
   idempotently inserts the example user;
5. runs `go run . serve`.

The service is ready when both health probes succeed:

```sh
curl --fail http://localhost:8080/api/v1/healthz/liveness
curl --fail http://localhost:8080/api/v1/healthz/readiness
```

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
| `auth.jwt-secret` | `GOSTARTER_AUTH_JWT_SECRET` |
| `logging.level` | `GOSTARTER_LOGGING_LEVEL` |
| `logging.format` | `GOSTARTER_LOGGING_FORMAT` |
| `webserver.address` | `GOSTARTER_WEBSERVER_ADDRESS` |
| `webserver.port` | `GOSTARTER_WEBSERVER_PORT` |

The `.env` file is loaded by `make dev-run`; the application does not
automatically read dotenv files when invoked directly. To run commands
manually, export the file first:

```sh
set -a
. ./.env
set +a
go run . seed
go run . serve
```

Command flags take precedence over these environment variables. See
`go run . serve --help` and `go run . seed --help` for every supported setting.

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

Both `seed` and `serve` apply pending migrations before accessing application
tables. Running `seed` repeatedly is safe: existing baseline records are
skipped. Migration SQL lives in `internal/platform/database/postgres/migrations/` and is
embedded in the binary.

## Troubleshooting

Inspect container and health-check state:

```sh
docker compose ps
docker compose logs postgres
```

If port 5432 is already in use, stop the conflicting local PostgreSQL instance
or change both the Compose port mapping and `GOSTARTER_DATABASE_URI`.

To validate the Compose model without starting containers:

```sh
docker compose config
```

For generated code, tests, linting, and release commands, use the targets listed
by `make help`.

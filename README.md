# Go Starter

[![CI](https://github.com/rfizzle/go-starter/actions/workflows/ci.yml/badge.svg)](https://github.com/rfizzle/go-starter/actions/workflows/ci.yml)
[![Go 1.26.5](https://img.shields.io/badge/Go-1.26.5-00ADD8?logo=go&logoColor=white)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue)](LICENSE)

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
- testify and Testcontainers-Go for tests

[AGENTS.md](AGENTS.md) carries the repository rules — architecture boundaries,
engineering and testing policy, and the task-to-skill routing table — for both
humans and coding agents.

## Contents

- [What is included](#what-is-included)
- [Initialize a new project](#initialize-a-new-project)
- [Five-minute quickstart](#five-minute-quickstart)
- [Configuration](#configuration)
- [Core commands](#core-commands)
- [Architecture map](#architecture-map)
- [API and generation workflow](#api-and-generation-workflow)
- [Testing tiers](#testing-tiers)
- [Deployment notes](#deployment-notes)
- [AI-assisted development](#ai-assisted-development)
- [Contributing and security](#contributing-and-security)

## What is included

The production-oriented foundations are implemented and wired into `serve`:
early configuration validation, structured logging, graceful shutdown,
PostgreSQL pooling, explicit serialized migrations, JWT validation, health
endpoints, Prometheus metrics, request limits, configurable TLS, and optional
OTLP export. This is a starter, not a production deployment you should ship
unchanged: choose deployment-specific secrets, CORS origins, connection limits,
TLS termination, monitoring, backup, and availability policies first.

The user domain and configurable development seed record are examples intended
to be replaced or extended. The committed page under `web/build/` is a
framework-neutral placeholder, not a supported frontend toolchain; see
[`web/README.md`](web/README.md) for the embedding boundary. Login rate limiting
is active by default and process-local: each replica enforces its own limit.
The service ignores forwarding headers unless the immediate proxy is included
in `rate-limit.trusted-proxies`; configure only proxy CIDRs operated by your
deployment. OpenTelemetry export is inactive unless explicitly enabled and
given access to an OTLP collector.

## Initialize a new project

Run the initializer once in a clean clone before making application-specific
changes. It updates the Go module and imports, command and release names,
environment prefix, container metadata, OpenAPI metadata, examples, and
documentation:

```sh
make init \
  module=example.com/acme/orders \
  app-name='Acme Orders' \
  binary=orders-api \
  env-prefix=ORDERS \
  image=registry.example.com/acme/orders-api \
  description='Order processing API'
```

`make init` wraps `go run ./tools/initialize` and reports which value is
missing before invoking it; the underlying command accepts the same values as
`--module`, `--app-name`, `--binary`, `--env-prefix`, `--image`, and
`--description` flags.

The module, application name, binary name, environment prefix, and untagged OCI
image name are required; the description defaults to `<application name>
service`. Inputs are validated before any file is changed. The initializer only
rewrites Git-tracked, hand-written regular files: it skips `.git`, generated
ogen and sqlc output, the license attribution, and untracked files.

For safety, initialization fails when the worktree has staged, unstaged, or
untracked changes. As its final step the initializer deletes its own sources
(`internal/templateinit/` and `tools/initialize/`), so a derived project carries
no one-shot code and `make init` cannot run twice. There is intentionally no
force flag. After a successful run, review the diff, regenerate from the OpenAPI
and SQL sources, and validate the renamed project:

```sh
make generate-check
make check
make build
make release-snapshot
```

## Five-minute quickstart

You need Git, Go 1.26.5, Make, and Podman with a Compose provider (or another
way to run PostgreSQL).

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
configured development user (email and password from `.env`), and serves the
application. In another terminal, probe the service:

```sh
curl --fail http://localhost:8080/api/v1/healthz/liveness
curl --fail http://localhost:8080/api/v1/healthz/readiness
curl --fail http://localhost:8080/metrics
```

Use the seeded email and password to exercise the
authentication slice:

```sh
curl --fail-with-body \
  --header 'Content-Type: application/json' \
  --data '{"email":"admin@example.com","password":"<seed-password>"}' \
  http://localhost:8080/api/v1/login

# Copy the token field from the login response.
curl --fail-with-body \
  --header 'Authorization: Bearer <token>' \
  http://localhost:8080/api/v1/me
```

Login returns the same 401 response for an unknown email and a wrong password.
Successful responses carry a JWT in the JSON `token` field; `/me` accepts it in
the standard bearer header. Exceeding the configured per-client login limit
returns 429 with `Retry-After`.

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
search path is configured. For values used by `migrate`, `serve`, and `seed`,
precedence is:

1. command-line flags;
2. `GOSTARTER_*` environment variables;
3. the file passed with `--config`;
4. command defaults.

Nested keys replace dots and dashes with underscores. For example,
`database.uri` maps to `GOSTARTER_DATABASE_URI`, and
`webserver.read-header-timeout` maps to
`GOSTARTER_WEBSERVER_READ_HEADER_TIMEOUT`. `--verbose` is a shortcut that
forces debug logging. Run `./bin/gostarter migrate --help`,
`./bin/gostarter serve --help`, or `./bin/gostarter seed --help` for the
complete command-specific settings.

Configuration is validated before network or database resources are opened.
In particular, `serve` requires a PostgreSQL URI and a JWT secret of at least
32 bytes. The `seed` command runs only in the development environment, requires
`seed.email` and `seed.password`, and never prints the password. Secrets are
never included in validation errors or logs.

## Core commands

| Command | Purpose |
| --- | --- |
| `make help` | List supported targets. |
| `make init` | Rewrite the template identity once; see [Initialize a new project](#initialize-a-new-project). |
| `make tools` | Install the repository-pinned development tools. |
| `make env` | Report installed and required tool versions. |
| `make build` | Build `bin/gostarter` without changing source files. |
| `make check` | Run generation drift, formatting, vet, lint, and test checks. |
| `make test` | Alias for `make test-unit`. |
| `make test-unit` | Run tests that do not require external services. |
| `make test-integration` | Run integration tests with disposable external services. |
| `make test-all` | Run unit and integration test tiers. |
| `make race` | Run the default Go suite with the race detector. |
| `make cover` | Run unit and integration tests; report and gate maintained-code coverage. |
| `make generate` | Regenerate ogen and sqlc output. |
| `make generate-check` | Fail if committed generated code is stale. |
| `make architecture-check` | Verify module, domain, platform, and transport boundaries. |
| `make docs-check` | Verify documentation links, commands, paths, and Go version. |
| `make skills-check` | Validate the agent skill library and its discovery links. |
| `make agents-setup` | Regenerate the relative `.claude/` skill and agent links. |
| `make dev-up` | Start local PostgreSQL and wait for health. |
| `make dev-run` | Start PostgreSQL, migrate, seed, and serve using `.env`. |
| `make dev-down` | Stop local services without deleting database data. |
| `make dev-reset` | Delete local database data and start PostgreSQL fresh. |
| `make release-snapshot` | Build local release archives in `dist/`. |
| `make container` | Build the current container image with Podman. |
| `make container-check` | Validate the release and container packaging configuration. |
| `make vuln` | Scan Go dependencies with `govulncheck`. |
| `make secrets-check` | Scan the full git history for committed secrets with `gitleaks`. |
| `make ci` | Run the read-only validation workflow: `make check`, `make secrets-check`, and `make vuln`. |

`make tools` reads exact versions from `scripts/tool-versions.env`. Set
`TOOL_BIN` to install into an isolated directory; otherwise tools are installed
to `GOBIN` or the first `GOPATH` bin directory. It is the shared entrypoint for
local and CI setup, and `make env` compares the binaries on `PATH` with the same
declarations.
Mutating maintenance commands such as `make fmt`, `make generate`, and
`make tidy` are intentionally explicit.

## Architecture map

| Path | Responsibility |
| --- | --- |
| `main.go`, `cmd/` | Process entrypoint and Cobra commands. |
| `api/openapi.yml` | Source of truth for the HTTP API. |
| `internal/api/v1/` | ogen-generated transport code; never edit directly. |
| `internal/app/` | Composition root, dependency assembly, and application lifecycle. |
| `internal/domain/` | Infrastructure-free business data, repository contracts, and business errors. |
| `internal/modules/` | Business capabilities; `user/` is the example PostgreSQL-backed module. |
| `internal/platform/` | Technical infrastructure: config, database, HTTP server and middleware, observability, and rate limiting. |
| `internal/controller/` | Hand-written generated-transport-to-domain mapping. |
| `internal/auth/` | Authentication and generated security-handler integration. |
| `internal/cli/` | Shared CLI lifecycle and seed support. |
| [`web/`](web/README.md) | Minimal embedded frontend placeholder and its replacement contract. |

See the [documentation index](docs/README.md), the
[module/platform boundary decision](docs/architecture/module-platform-boundaries.md),
and the [current architecture notes](docs/architecture/current-state.md) for
deeper context; [AGENTS.md](AGENTS.md) holds the engineering policy.

## API and generation workflow

API changes start in `api/openapi.yml`. Run `make generate`, then implement any
changed generated interfaces in `internal/controller/`. SQL changes start in
`internal/platform/database/postgres/queries/` or a new sequential migration, followed by
`make generate`. Never hand-edit `internal/api/v1/oas_*_gen.go` or the
sqlc-generated files in `internal/platform/database/postgres/db/`.

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
- Full unit/package suite: `make test-unit` (also available as `make test`).
- Disposable PostgreSQL integration suite: `make test-integration`.
- Both tiers: `make test-all`.
- Static and generated-code gates: `make check`.
- Concurrency-sensitive changes: `make race`.
- Coverage report and gate: `make cover` (fails below the floor in `scripts/coverage-baseline.env`).
- Dependency vulnerability scan: `make vuln`.
- Committed-secret scan: `make secrets-check`.

Integration tests use the
[PostgreSQL test harness](internal/platform/database/postgrestest/README.md),
which starts an isolated container through the Podman API socket and skips
locally with a clear diagnostic when no container engine is reachable. CI sets
`GOSTARTER_REQUIRE_CONTAINERS=1` and treats startup failure as a failed
integration run. See the
[development environment guide](docs/development/environment.md) for the
one-time Podman socket setup.

The testing policy in [AGENTS.md](AGENTS.md) describes project
conventions. Unit tests use narrow fakes where appropriate; integration and
application-lifecycle tests use disposable PostgreSQL containers for behavior
that mocks cannot prove.

## Deployment notes

`make release-snapshot` uses `.goreleaser.yaml` to produce explicit amd64 and
arm64 archives, checksums, archive SBOMs, and nonpublishing container images.
Its image step is the repository's one retained Docker dependency: GoReleaser's
`dockers_v2` pipeline drives Docker Buildx directly and has no Podman backend.
On a Podman-only host, build everything except the images with
`goreleaser release --clean --auto-snapshot --skip=publish,announce,docker`.
`make container` builds only the local architecture from
`build/docker/Dockerfile` with rootless Podman; both final images run as a
non-root user and invoke `/app serve`.

Pushing a valid `vX.Y.Z` semantic-version tag (optionally with a prerelease
suffix) runs the least-privilege release workflow. Tag pushes do not trigger
`ci.yml`, so the release job runs `make ci` itself and refuses to publish when
those gates fail. Build metadata suffixes are not accepted because `+` is not
valid in an OCI image tag. GoReleaser generates
the changelog and GitHub Release assets and publishes the `linux/amd64` and
`linux/arm64` image to
`ghcr.io/rfizzle/go-starter` with the release tag. Stable releases also update
the `latest` image tag. Git tags are the only release-version source; untagged
local container builds use a commit-derived `0.0.0-dev.*` version. The OpenAPI
`info.version` describes the API contract and is intentionally independent.

The release job then attests build provenance for the published archives and
the container image, so a consumer can verify what built an artifact:

```sh
gh attestation verify <downloaded-archive>.zip --repo <owner>/<repository>
gh attestation verify oci://ghcr.io/rfizzle/go-starter:v1.2.3 --repo <owner>/<repository>
```

The [CI contract](docs/development/ci-contract.md) documents every workflow gate.

The Podman build also supports multi-platform manifest lists and an OCI
archive export. For example, after creating `dist/`, this builds both release
platforms and writes an OCI archive without publishing it (the non-native
platform needs binfmt emulation via `qemu-user-static`):

```sh
CONTAINER_PLATFORMS=linux/amd64,linux/arm64 \
  CONTAINER_ARCHIVE=dist/starter-image.tar \
  make container
```

Both base-image tags are pinned to multi-platform index digests. Dependabot
checks the Dockerfile weekly. Review a proposed tag and digest together, confirm
that the Go builder version still agrees with `go.mod`, then run
`make container-check` and `make container` before merging the update.

Run `/app migrate up` as a deployment job before starting application replicas.
The command applies pending embedded migrations under a PostgreSQL session
advisory lock, so concurrent attempts serialize. An init container is also
safe, but runs once per pod; a single job provides a clearer independent schema
gate. Startup migration is disabled by default. A deployment that deliberately
chooses that mode can set `GOSTARTER_DATABASE_AUTO_MIGRATE=true`; migration
still runs under the lock and any failure exits before the HTTP listener starts.

Provide configuration through deployment-managed environment variables or an
explicit config file, and supply PostgreSQL separately. Plain HTTP is the
default because TLS is normally terminated by a trusted ingress or load
balancer; set both TLS file options only when the process terminates TLS itself.
Profiling is disabled by default and should remain private when enabled.

A production Compose stack and deployment manifests are not included. Validate
the image, health probes, secret delivery, CORS policy, migration job,
observability, and rollback strategy in your target platform.

## AI-assisted development

[AGENTS.md](AGENTS.md) is the entry point for coding agents: it states the
architecture boundaries, the engineering and testing policy, and which skill to
load for a given task. `CLAUDE.md` includes it, so Claude Code picks it up
automatically.

The skills themselves ship with the template in
[`.agents/skills/`](.agents/skills/README.md), a client-neutral library that
follows the [Agent Skills specification](https://agentskills.io/specification).
`.claude/skills/` and `.claude/agents/` hold relative symlinks into that library
and into `docs/ai/prompts/`, so discovery works in a fresh clone. Regenerate the
links with `make agents-setup` after adding a skill; `make skills-check` runs as
part of `make check` and fails on an invalid skill or a broken link. It needs
Node.js on `PATH`.

`.claude/settings.json` is the committed permission baseline shared by the
project. Keep personal overrides in the ignored `.claude/settings.local.json`.

## Contributing and security

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow and
[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for participation expectations.
Notable changes are summarized in [CHANGELOG.md](CHANGELOG.md). Report
suspected vulnerabilities according to [SECURITY.md](SECURITY.md), without
placing sensitive details in a public issue.

Go Starter is available under the [MIT License](LICENSE).

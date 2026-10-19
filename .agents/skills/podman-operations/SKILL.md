---
name: podman-operations
description: "Operate Go Starter's container workflows with Podman: local PostgreSQL through podman compose (dev-up, dev-run, dev-reset), testcontainers-backed integration tests over the Podman API socket, rootless image builds with make container, multi-architecture manifests and OCI archives, and the GoReleaser release boundary. Use when preparing a Podman host, debugging compose-provider or API-socket failures, running the integration tier, building or exporting container images, or translating Docker habits to this repository. Not for writing repository or SQL code (golang-database) or for test design (golang-testing)."
metadata:
  author: rfizzle
  standard: agentskills.io
---

# Podman Operations

Podman is this repository's container engine, rootless, and its
Docker-compatible API socket is what testcontainers talks to. The one retained
Docker dependency is GoReleaser's image pipeline; see
[Release boundary](#release-boundary-the-retained-docker-dependency).

## Host requirements

- Go at the version `go.mod` declares, and Make.
- The pinned development tools: `make tools` installs the exact versions in
  `scripts/tool-versions.env` (into `GOBIN` or the first `GOPATH` bin) and
  `make env` verifies them. To keep them isolated, pick a directory and put it
  first on `PATH`:

  ```sh
  TOOL_BIN=/absolute/path/to/project-tools make tools
  PATH="/absolute/path/to/project-tools:${PATH}" make env
  ```

  ogen is not among them: it is a module dependency run by `go generate`.
- Podman 4.7 or newer (`podman --version`).
- A Compose provider: `podman compose` delegates to `podman-compose` or a
  Compose v2 `docker-compose` binary on `PATH`. The Make targets pass Compose v2
  flags (`up --detach --wait`), so the provider must support them.
- Linux integration tests need the rootless API socket:

  ```sh
  systemctl --user enable --now podman.socket
  ```

- macOS runs Podman inside a managed VM; create and start it once with
  `podman machine init` and `podman machine start`. The machine exposes the API
  socket automatically.
- Cross-architecture image builds additionally need binfmt emulation
  (`qemu-user-static` on Linux; the Podman machine already ships it).

Verify a host with `make env` (tool presence) and `podman info` (engine and
socket state).

## Local development services

`make dev-up`, `make dev-run`, `make dev-down`, and `make dev-reset` wrap
`podman compose` around the committed `compose.yaml` PostgreSQL service: the
pinned `postgres:17.5-alpine` image, bound to `127.0.0.1:5432`, data in a named
volume, and development-only `gostarter` credentials that match
`.env.example`. `dev-down` preserves the volume; `dev-reset` is the only target
that deletes it.

First run:

```sh
cp .env.example .env
make dev-run
curl --fail http://localhost:8080/api/v1/healthz/liveness
curl --fail http://localhost:8080/api/v1/healthz/readiness
```

`dev-run` starts the service, waits for `pg_isready`, exports `.env`, then
runs `migrate up`, `seed` and `serve`. Only `dev-run` reads `.env`; to run a
command by hand, export it first:

```sh
set -a; . ./.env; set +a
go run . serve
```

Diagnose the service with:

```sh
podman compose ps
podman compose logs postgres
podman compose config
```

## Integration tests (testcontainers)

The [postgrestest harness](../../../internal/platform/database/postgrestest/README.md)
starts one disposable PostgreSQL container per test binary through
testcontainers, which speaks the Docker-compatible API. Point it at Podman:

```sh
export DOCKER_HOST="unix://$(podman info --format '{{.Host.RemoteSocket.Path}}')"
export TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED=true
make test-integration
```

- `DOCKER_HOST` is required whenever testcontainers cannot discover the Podman
  socket on its own; exporting it explicitly is the reliable default.
- Ryuk, the testcontainers reaper container, needs the privileged flag under
  rootless Podman.
- With no reachable engine, integration tests skip with a diagnostic. Setting
  `GOSTARTER_REQUIRE_CONTAINERS=1` (CI does) turns that skip into a failure.

## Image builds

`make container` runs `scripts/container-build.bash`, which builds
`build/docker/Dockerfile` with `podman build`. The Dockerfile format and
`.dockerignore` are OCI-ecosystem standards that Podman consumes natively; no
file needs a Docker daemon.

- Default: a single image for the local architecture, tagged with the commit,
  the derived version, and `latest`.
- `CONTAINER_PLATFORMS=linux/amd64,linux/arm64` builds both platforms into a
  local manifest list (requires binfmt emulation for the non-native one).
- `CONTAINER_ARCHIVE=dist/starter-image.tar` additionally exports an OCI
  archive via `podman save` (single platform) or `podman manifest push --all`
  (manifest list).
- Publish manually with `podman push <image> docker://<registry>/<name>:<tag>`
  or `podman manifest push --all` for manifest lists.

`make container-check` validates the GoReleaser packaging configuration; the
real `make container` build is what proves the Dockerfile against the tree.

## Release boundary (the retained Docker dependency)

GoReleaser's `dockers_v2` image pipeline drives Docker Buildx directly and has
no Podman backend, so release image packaging in `.github/workflows/release.yml`
and a full local `make release-snapshot` are the only steps that still require
Docker. On a Podman-only host, build the archives, checksums, and SBOMs without
the image step:

```sh
goreleaser release --clean --auto-snapshot --skip=publish,announce,docker
```

Revisit this boundary if GoReleaser gains Podman support.

## Gotchas

- "no compose provider found" from `podman compose` means neither
  `podman-compose` nor `docker-compose` is on `PATH`; install one.
- A `DOCKER_HOST` pointing at a stopped socket makes integration tests report
  an unavailable provider; restart `podman.socket` or unset the variable.
- Repeated manual `podman manifest` work appends entries to an existing list;
  `podman manifest rm` first. `scripts/container-build.bash` already does this.
- `podman system reset` deletes every image, container, and volume, including
  the development database volume; prefer `make dev-reset` for a clean database.

## Verification

`make env`, then the workflow-relevant subset of `make dev-up`,
`make test-integration`, and `make container`. Finish any repository change
with `make check`.

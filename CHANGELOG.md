# Changelog

All notable changes to Go Starter are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
intends to follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
once it reaches 1.0.

GoReleaser also generates GitHub Release notes from the commit log on every
tag, filtering `docs:`, `test:`, `chore:`, and merge commits. That generated
list is the per-release record of what shipped; this file is the curated,
human-facing summary. Keep an entry here for anything a consumer of the
template must act on — a breaking change, a new required setting, a migration,
or a removed target.

Add unreleased entries under the heading below, and move them into a new
version section when the `vX.Y.Z` tag is pushed.

## [Unreleased]

### Added

- `podman-operations` skill documenting the Podman host requirements, compose
  provider, testcontainers socket setup, image builds, and the GoReleaser
  release boundary.
- `CONTAINER_ARCHIVE` build-script option that exports the built image or
  manifest list as an OCI archive.

### Changed

- Podman replaces Docker as the container engine for every developer-facing
  workflow: `dev-up`/`dev-run`/`dev-down`/`dev-reset` use `podman compose`,
  `make container` (previously `make docker`) builds images with rootless
  `podman build`, and the integration tier reaches Podman through its
  Docker-compatible API socket. GoReleaser's `dockers_v2` release image
  packaging is the one retained Docker dependency.
- Integration tests now require `GOSTARTER_REQUIRE_CONTAINERS=1` (previously
  `GOSTARTER_REQUIRE_DOCKER=1`) to make a missing container engine fatal.
- `scripts/docker-build.bash` is now `scripts/container-build.bash`;
  multi-platform builds produce a Podman manifest list and the
  `CONTAINER_OUTPUT` Buildx option is replaced by `CONTAINER_ARCHIVE`.
- `make container-check` validates the GoReleaser packaging configuration;
  the Dockerfile itself is proven by the real `make container` build in CI.

### Deprecated

### Removed

- `make docker` target and the Docker/Buildx prerequisite for local
  development. Install Podman 4.7+ with a compose provider instead.

### Fixed

### Security

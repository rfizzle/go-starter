# CI Contract

This document states what each GitHub Actions workflow enforces, so a change
author can predict which gate will catch a mistake and reproduce it locally.

Every automated gate is a Make target. CI installs the pinned toolchain and
then runs the same commands a contributor runs, which is why
`make ci` locally is a faithful preview of the tag-push gate.

## Workflows

| Workflow | Triggers | Enforces |
| --- | --- | --- |
| `.github/workflows/ci.yml` | pull request, push to `master` | Repository gates, secret scan, vulnerability scan, race suite, coverage floor, local build, container build, release snapshot, and a cross-platform compile |
| `.github/workflows/codeql.yml` | pull request, push to `master`, weekly | CodeQL static analysis of the Go tree; results land in the repository's code-scanning alerts |
| `.github/workflows/release.yml` | push of a `vX.Y.Z` tag | Tag validity, the full `make ci` gate, GoReleaser configuration, publication, and build provenance |

### ci.yml

The `Validate` job runs, in order: `make check`, `make secrets-check`,
`make vuln`, `make race`, `make cover` (with `GOSTARTER_REQUIRE_CONTAINERS=1`,
so a container failure is an error rather than a skip), `make build`,
`make container-check`, `make container`, and `make release-snapshot`.

Before `make cover`, the job enables the runner's rootless Podman API socket
and exports `DOCKER_HOST` (plus the privileged Ryuk flag) so the
testcontainers-backed integration tier talks to Podman.

`make check` itself is the fast read-only bundle: `make architecture-check`,
`make docs-check`, `make skills-check`, `make generate-check`,
`make fmt-check`, `make lint`, and `make test`.

The job checks out full history because `make secrets-check` scans commits; a
shallow clone would inspect only the tip and pass silently.

The `Cross-platform build` job compiles the tree on `macos-latest` and
`windows-latest`. GoReleaser ships darwin and windows binaries, but every other
gate runs on Linux, so this is the only step that would catch a build-tag or
platform-specific break before a tag turns it into a broken release. It runs
`go build ./...` rather than `make build`: `make` is not part of the default
Windows runner image, and proving the tree compiles is the whole point. Tests
stay Linux-only because the integration tier needs Podman.

There is no Go N-1 matrix leg. `go.mod` requires an exact patch release, so a
runner pinned to an older minor version would download the required toolchain
and analyze the same compiler anyway. Lowering the `go` directive to make an
N-1 leg meaningful is a language-support decision, not a CI one.

### release.yml

A tag push does not trigger `ci.yml`, so the release job re-runs the read-only
gates itself through `make ci` and refuses to publish when they fail.

Release image packaging is the repository's one retained Docker dependency:
GoReleaser's `dockers_v2` pipeline drives Docker Buildx directly and has no
Podman backend, so this job logs in to the registry with Docker on the runner.
Every developer-facing container workflow uses Podman. After
`goreleaser release`, two provenance attestations are generated with
`actions/attest-build-provenance`:

- the release archives and their SBOMs, identified through
  `dist/checksums.txt`; and
- the published multi-platform image, identified by the index digest that
  GoReleaser writes to `dist/digests.txt` and pushed to the registry alongside
  the image.

Both are Sigstore keyless signatures backed by the workflow's OIDC identity,
which is why the job requests `attestations: write` and `id-token: write`.
Consumers verify them with the GitHub CLI:

```sh
gh attestation verify <downloaded-archive>.zip --repo <owner>/<repository>
gh attestation verify oci://ghcr.io/rfizzle/go-starter:v1.2.3 --repo <owner>/<repository>
```

## Secret scanning

`make secrets-check` runs the pinned `gitleaks` binary over the full commit
history. Findings are redacted in output so a detected value never reaches a
public log; the report still names the commit, file, and line.

`.gitleaks.toml` extends the upstream default rule set and records the reviewed
exceptions. It currently allowlists Markdown under `.agents/skills/`, which
ships illustrative bad-example credentials as documentation.

This is a repository-portable gate rather than a GitHub setting, so a clone
keeps it regardless of the account or plan it lands in. It complements, and
does not replace, enabling GitHub push protection on the hosted repository.

## Pinned tool versions

`scripts/tool-versions.env` declares the exact version of every development
tool, and `make tools` installs them for both local use and CI.

**Dependabot does not watch this file.** The configured ecosystems are Go
modules, GitHub Actions, Dockerfiles, and Compose files; a `go install` target
declared in a shell environment file is invisible to all of them. Review these
pins periodically — checking each tool's latest release, bumping the version,
and running `make env` followed by `make ci` — and treat a pin that is many
releases behind as a finding rather than a steady state.

Action versions in the workflow files are pinned to full commit SHAs with a
trailing version comment. Dependabot proposes updates for those and rewrites
both the SHA and the comment.

## Reproducing a CI failure locally

```sh
make tools
make env
make ci
```

Add `make race` and `make cover` when the failure is in those steps; the
integration tier needs a reachable Podman socket (see the
[environment guide](environment.md)). `make container-check`,
`make container`, and `make release-snapshot` reproduce the container and
release-packaging steps.

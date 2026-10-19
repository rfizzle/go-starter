# Deployment

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

The [CI contract](ci-contract.md) documents every workflow gate.

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

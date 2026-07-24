# Starter Repo Hardening Review — 2026-07-03

> **Status: historical snapshot.** This review records the repository as it
> existed on 2026-07-03. Its paths, versions, and checklist state are preserved
> as review evidence and are not current instructions. The ordered
> `Go-Starter-TODO.md` backlog superseded it; GS-01 through GS-29 implemented
> the accepted hardening scope, while unchecked ideas that were not selected
> remain non-commitments. Use the root `README.md`, `AGENTS.md`, and
> `docs/architecture/` for current guidance.

A full review of `go-starter` as a **reusable template**. The goal is not just to
fix this repo, but to solidify patterns that carry over to every future Go
project scaffolded from it. Findings are grouped by theme and written as a
checklist to work through.

Legend: 🔴 critical / correctness · 🟠 high-value · 🟡 polish · 💡 pattern to solidify

Each item cites `file:line` where relevant. Check items off as completed. Items
are ordered roughly by impact within each section.

---

## 0. Bugs & Correctness (do these first)

- [x] 🔴 **Dockerfile defaults to Go 1.19 but the module requires 1.25.7 — the image cannot compile.**
  `build/docker/Dockerfile:5` sets `ARG GO_VERSION=1.19`, `go.mod:3` says `go 1.25.7`,
  and `scripts/docker-build.bash` never passes `--build-arg GO_VERSION=...`. `make docker`
  will fail. Fix: bump the default to `1.25` (or read it from `go.mod`) and/or pass it
  from the build script.
- [x] 🔴 **Version source-of-truth is split three ways.** `.version` = `0.0.1`,
  `package.json` = `0.2.0`, and `scripts/docker-build.bash` tags images from
  `package.json` while `scripts/sync-version.bash` reads `.version`. Pick ONE source
  (recommend a git tag → `.version`) and derive everything else from it. See §3.
- [x] 🟠 **`middleware.Logger` path-filter loop is fragile for >1 path.**
  `internal/platform/middleware/logger.go:18-23` returns (skips logging) on the *first* path that
  doesn't match, so it only behaves correctly because a single pattern (`/api/*`) is
  passed in `internal/platform/web/webserver.go:60`. With multiple patterns it would skip a
  request that matches the 2nd pattern but not the 1st. Rewrite as "log if ANY pattern
  matches" (loop, set a `matched` bool, then decide).
- [x] 🟠 **`pgxpool` is created with all-default settings.** `internal/app/app.go:49`
  uses `pgxpool.New(ctx, uri)` — no `MaxConns`, `MinConns`, `MaxConnLifetime`,
  `MaxConnIdleTime`, or health-check interval. Under load the pool defaults
  (`4 × NumCPU`) are often wrong. Switch to `pgxpool.ParseConfig` + `NewWithConfig`
  and expose the knobs via config. See §5.
- [x] 🟡 **`pool.Close()` on shutdown has no timeout** (`internal/app/app.go:57`). If a
  connection hangs, shutdown can block indefinitely. Bound it.

---

## 1. HTTP Server Hardening

The graceful-shutdown implementation is genuinely good — signal handling via
`signal.NotifyContext`, readiness flipped false *before* draining, buffered error
channel, fresh shutdown context. Keep that as the reference pattern. Gaps:

- [x] 🟠 **Add `ReadHeaderTimeout`** to the `http.Server` (`internal/platform/web/webserver.go:89-96`).
  Without it the server is exposed to Slowloris header attacks. Suggest 5s.
- [x] 🟠 **Set `MaxHeaderBytes`** explicitly (e.g. `1 << 20`) at the same spot.
- [x] 🟠 **Replace chi's `Recoverer`** (`webserver.go:59`) with a custom panic-recovery
  middleware that logs to the zap logger (with request-id + stack) and increments a
  panic metric. chi's default writes to stdlib `log`, so panics are invisible in the
  structured JSON logs.
- [x] 🟡 **Add a request body size limit** (`http.MaxBytesHandler`, e.g. 10MB) as
  middleware to cap memory from oversized POSTs.
- [x] 🟡 **Document the TLS story.** The server assumes TLS is terminated by a reverse
  proxy. State that explicitly and, optionally, add a `WithTLS`/`MinVersion` option for
  standalone deployments.
- [x] 💡 **Pattern:** keep the functional-options constructor (`With…`) — it's clean and
  worth carrying forward. Add the new hardening knobs as options with safe defaults.

---

## 2. Observability

Structured logging (zap, env-aware JSON/console), `/metrics`, request IDs, and
config-gated pprof are all present and well done. The big gap:

- [x] 🟠 **OpenTelemetry is imported but never initialized.** The generated ogen code and
  `go.mod` reference otel, but no `TracerProvider`/`MeterProvider` is created in
  `internal/app/app.go`, and no exporter (OTLP) is configured — so spans/metrics are
  dropped. Either (a) wire up an OTLP exporter with a `WithOTel` config block, or
  (b) if it's aspirational, document it as "not yet wired" so it isn't mistaken for
  working. Decide and make it honest.
- [x] 🟠 **Export DB pool metrics.** Periodically record `pool.Stat()` (acquired / idle /
  total / wait count) so connection exhaustion is visible before queries start timing
  out.
- [x] 🟡 **Enrich request logs** with response size (`ww.BytesWritten()`) and request
  size (`r.ContentLength`) in `internal/platform/middleware/logger.go`.
- [x] 🟡 **Log readiness state transitions** so probe flaps are traceable.
- [x] 💡 **Pattern:** a single `internal/platform/observability` (or `internal/telemetry`) package
  that owns logger + tracer + meter provider construction and returns a `func()` cleanup.
  Carries over to every project.

---

## 3. Release Cycle & Versioning

- [x] 🔴 **Unify the version source** (see §0). Recommended flow: git tag `vX.Y.Z` is the
  truth; CI writes `.version`; `sync-version.bash` propagates into the OpenAPI spec.
  Remove the `package.json` version from the release path.
- [x] 🟠 **Add [goreleaser](https://goreleaser.com/)** (`.goreleaser.yaml`) to replace the
  hand-rolled `scripts/build.bash` cross-compile: builds Linux/macOS/Windows +
  arm64/amd64, checksums, changelog from conventional commits, GitHub Release, and
  Docker image publishing. This is the single biggest reuse win.
- [ ] 🟠 **Generate a `CHANGELOG.md`** from conventional commits (goreleaser or git-cliff).
  Commit history already follows `feat:`/`fix:`/`refactor:` — lean into it.
- [ ] 🟡 **Inject build metadata via ldflags in the Docker build too** (version, commit,
  date) so `cmd/version.go` output is correct for container images.
- [ ] 💡 **Pattern:** ldflags-injected `Build*` vars + a `version` subcommand is already a
  good pattern here — keep it, just feed it from a single source.

---

## 4. CI/CD & Automation (biggest structural gap)

There is **no `.github/` directory at all** — everything relies on running `make`
locally. For a template, baked-in CI is the highest-leverage addition.

- [ ] 🔴 **Add `.github/workflows/ci.yml`**: on push/PR, run `make ci` (test + fmt-check +
  lint), with the Go version pinned from `go.mod`. Cache the module/build cache.
- [ ] 🟠 **Add `govulncheck`** to CI (`golang.org/x/vuln/cmd/govulncheck ./...`) and a
  `make vuln` target. Cheap, high-signal supply-chain check.
- [ ] 🟠 **Add `.github/dependabot.yml`** (or Renovate) for `gomod`, `docker`, and
  `github-actions` ecosystems, with grouped minor/patch updates.
- [ ] 🟠 **Add `.github/workflows/release.yml`** triggered on tag push → runs goreleaser.
- [ ] 🟡 **Add CodeQL** (`.github/workflows/codeql.yml`) and **Trivy** image scanning
  (SARIF upload) for defense-in-depth.
- [ ] 🟡 **Generate an SBOM** on release (syft/goreleaser) and attach it to the release.
- [ ] 🟡 **Add `SECURITY.md`** with a vulnerability-reporting policy.
- [ ] 🟡 **Coverage gate:** add a `make cover` target and re-enable `.codecov.yaml`
  enforcement (currently `project: false / patch: false`, i.e. disabled).
- [ ] 💡 **Pattern:** ship these workflows *in the template* so every new project starts
  with CI, dependency updates, and release automation on day one.

---

## 5. Storage & Data Layer

Architecture is strong: clean `storage.Client`/`Transaction` interfaces, sentinel
errors (`internal/platform/database/errors.go`), goose migrations, and a deliberate
**sqlc (static, type-safe queries) + squirrel (dynamic filters)** split — that
hybrid is *not* redundant; document it so it's not "cleaned up" by mistake.

- [ ] 🟠 **Expose pgxpool config** (see §0) via `internal/platform/config` + `cmd/serve.go` flags:
  `max-conns`, `min-conns`, `max-conn-lifetime`, `max-conn-idle-time`.
- [ ] 🟠 **Redact SQL/args in debug logs.** `internal/platform/database/postgres/postgres.go` and
  `transaction.go` log `zap.Any("args", args)` — bind parameters can contain PII or
  secrets. Gate behind an explicit "log SQL" flag and/or redact.
- [ ] 🟡 **Add a goose advisory `SessionLocker`** for safe concurrent multi-instance
  startup migrations (`internal/platform/database/postgres/migration.go` already notes the race).
- [ ] 🟡 **Wrap raw errors with operation context** in the low-level `_exec`/`_query`
  helpers so failures are traceable to a call site.
- [ ] 💡 **Pattern:** document the sqlc-vs-squirrel decision rule in
  `docs/development/database.md` (static → sqlc, dynamic filters → squirrel).

---

## 6. Configuration

Viper + Cobra with flag > env > default precedence and env-aware logging is solid
and 12-factor-friendly. Gaps:

- [ ] 🟠 **Add a `config.Validate()` method** (`internal/platform/config/config.go` has none) called in
  `cmd/serve.go` before app init: require `Database.URI` and `Auth.JWTSecret`, validate
  port range, etc. Fail fast at startup instead of deep in a request.
- [ ] 🟡 **Redact secrets** (`Database.URI` password, `Auth.JWTSecret`) anywhere config is
  logged or printed.
- [ ] 🟡 **Ship a `.env.example`** and document config-file loading (YAML/JSON) for
  non-container deployments.
- [ ] 💡 **Pattern:** consider `go-playground/validator` struct tags (`validate:"required"`)
  so validation is declarative and reusable across projects.

---

## 7. Testing

Excellent testing *docs* (`docs/guidelines/testing-strategy.md`: pyramid, TDD,
table-driven, 80-90% target) and 23 unit-test files using `pgxmock`. The gap is
that the strategy's higher tiers aren't implemented.

- [ ] 🟠 **Add integration tests with [testcontainers-go](https://golang.testcontainers.org/)**
  against a real Postgres, behind a build tag / `make test-integration`. Mocks alone
  don't catch SQL/migration errors.
- [ ] 🟠 **Add a coverage target** (`make cover`) and wire it into CI with a threshold.
- [ ] 🟡 **Add benchmarks** for hot paths (JWT validate, query builders); the docs
  describe them but none exist.
- [ ] 🟡 **Add E2E API tests** exercising the ogen handlers end-to-end (spin server,
  hit `/api/v1/...`).
- [ ] 💡 **Pattern:** a reusable `internal/testutil` (or `testsupport`) package with a
  testcontainers Postgres fixture + fixture factories — carries to every project.

---

## 8. Rate Limiting

`internal/platform/ratelimiter` has four algorithms (fixed window, sliding window, token
bucket, Redis skeleton) but is **not wired into the server** and has sharp edges.

- [ ] 🟠 **Mount the middleware** on the API mux (or document that it's opt-in). Today it's
  dead code from the server's perspective.
- [ ] 🟠 **The `RedisRateLimiter` is a non-functional skeleton.** Either implement it
  (atomic Lua INCR+EXPIRE) or remove it so nobody deploys thinking it's distributed.
  In-memory limiters are per-pod and won't hold in a multi-replica deploy — call this
  out.
- [ ] 🟡 **Bound memory** in the in-memory limiters (cap total keys / evict) to avoid
  unbounded growth under key-cardinality attacks.
- [ ] 🟡 **Harden client-IP extraction** — `X-Forwarded-For` is spoofable unless you trust
  a known proxy; prefer chi `RealIP` behind a trusted proxy or validate the hop.
- [ ] 🟡 **Add rate-limit metrics** (allow/deny counters).

---

## 9. Docker

Multi-stage build on distroless with a non-root user is a good baseline.

- [ ] 🔴 **Fix the Go version** (see §0) — highest priority.
- [ ] 🟠 **Add build-arg version/commit injection** so image builds stamp `cmd/version.go`.
- [ ] 🟡 **Multi-arch builds** (`docker buildx --platform linux/amd64,linux/arm64`) in
  `scripts/docker-build.bash` (or hand this to goreleaser).
- [ ] 🟡 **Pin base images by digest** for reproducibility.
- [ ] 🟡 Note: distroless has no shell, so a `HEALTHCHECK` CMD isn't straightforward —
  prefer Kubernetes liveness/readiness probes hitting the existing `/api/v1` health
  endpoints. Document this rather than adding a broken healthcheck.

---

## 10. Health Checks

Liveness (always-200) vs readiness (atomic flag) is correctly distinguished
(`internal/controller/health.go`) and flipped appropriately on startup/shutdown.

- [ ] 🟠 **Readiness should check dependencies.** It only reads the atomic flag; add a
  bounded `Store.Ping(ctx)` so a dead DB actually de-registers the pod from traffic.
- [ ] 💡 **Pattern:** a small health-aggregator (named checks → status) that both the
  readiness endpoint and a verbose `/healthz?verbose` can share.

---

## 11. Frontend / `web/` Scaffolding (framework-agnostic)

The Go side is already wired: `web/web.go` embeds `build/*` via `//go:embed`,
`AssetHandler` does SPA fallback to `index.html`, `internal/platform/middleware/static.go`
routes non-API/metrics/debug traffic to it, and `internal/app/app.go` passes it in
via `WithFileServer`. What's missing is the JS-side scaffolding — and it can be
added **without** picking React vs Svelte by standardizing on pnpm + Vite (Vite
supports React/Svelte/Vue/vanilla behind the same config).

- [ ] 🟠 **Add pnpm workspace**: `pnpm-workspace.yaml` at root + `web/package.json`. Pin
  the pnpm version via `packageManager` in `package.json` + Corepack.
- [ ] 🟠 **Add Vite as the neutral build tool** targeting `web/build/` (the existing embed
  dir) as `outDir`, so `go:embed` keeps working unchanged. Framework plugin (React/
  Svelte/etc.) is the *only* thing a downstream project swaps.
- [ ] 🟠 **Wire the frontend build into the Go build**: add `make web-install`, `make
  web-build`, `make web-dev` targets; have `make build` (and `scripts/build.bash`) run
  the web build before `go build` so assets are embedded. Guard on `pnpm` being present.
- [ ] 🟡 **Dev-server proxy**: add a `--web-dev` flag (or document a reverse proxy) so the
  Go server proxies to Vite's dev server (`:5173`) for HMR instead of serving embedded
  assets during development.
- [ ] 🟡 **Add Storybook** (`web/.storybook/`) sharing the Vite config for isolated
  component development.
- [ ] 🟡 **Add JS toolchain**: TypeScript, Prettier, ESLint, Vitest — all framework-neutral.
- [ ] 🟡 **Update `.gitignore`**: `web/node_modules/`, `web/build/` (generated),
  `web/dist/`, framework caches (`.svelte-kit/`, `.next/`), coverage. Keep a
  `web/build/.keep` or a placeholder so `//go:embed build/*` compiles on a clean checkout
  **before** the first frontend build (otherwise `go build` fails — important for a
  template).
- [ ] 🟠 **Guarantee `go build` works with no frontend built.** Because `web/web.go` embeds
  `build/*`, a fresh clone must still contain at least one file there (currently
  `index.html`, `sample.txt`). Document this invariant and/or generate a stub in the
  build so contributors who `rm -rf web/build` don't break compilation.
- [ ] 💡 **Pattern:** commit a minimal placeholder `build/` + a `web/README.md` explaining
  "choose your framework here, everything else is wired." This is the reusable win — the
  embed/serve/build pipeline is decided; only the UI library is deferred.

---

## 12. Cross-cutting patterns to solidify in the template

- [ ] 💡 **Pre-commit hooks**: add `.pre-commit-config.yaml` (or Husky, since
  `package.json` exists) running `make fmt lint` + conventional-commit check. `.git/hooks`
  is currently empty.
- [ ] 💡 **`internal/platform/observability` package** (logger+tracer+meter) — see §2.
- [ ] 💡 **`internal/testutil` package** (testcontainers fixture + factories) — see §7.
- [ ] 💡 **Config `Validate()` convention** — see §6.
- [ ] 💡 **Single version source + goreleaser + CI/release workflows** — see §3/§4. These
  are the pieces that make the *next* project start in a good place, not just this one.

---

## Suggested order of attack

1. **§0 bugs** (Docker Go version, version unification, logger loop, pool config) — these
   are correctness issues.
2. **§4 CI** + **§3 release** — unblock automation; highest reuse leverage.
3. **§1 HTTP hardening** + **§6 config validation** — cheap production-readiness.
4. **§2 observability wiring** + **§5 pool metrics/config**.
5. **§7 integration tests** + **§11 frontend scaffolding**.
6. Remaining 🟡 polish and 💡 pattern packages.

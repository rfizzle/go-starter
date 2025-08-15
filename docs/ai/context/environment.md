# Tech Environment (Standard, Reusable)

> **Scope:** Canonical local/tooling environment used by all agents and contributors.

---

## 1) Target Platform

- **Language:** Go (Golang)
- **Project layout:** Go Standard Project Layout
- **Primary dev machine:** macOS on Apple Silicon (M‑series)
- **OS/Arch:** `darwin/arm64`
- **Shell:** zsh (default macOS), bash acceptable
- **Package manager:** Homebrew (preferred)

**Apple Silicon notes**
- Prefer ARM64 binaries via Homebrew.
- Avoid CUDA‑only libs; use CPU/Metal builds when applicable.

---

## 2) Required Versions (defaults)

- **Go:** `>= 1.22` (use latest stable patch)
- **golangci-lint:** pinned in CI; local can follow latest
- **ogen:** latest stable (OpenAPI‑first)
- **Mock generator (optional):** **Uber** mockgen (`go.uber.org/mock/mockgen`)

> If a repo needs different versions, record them here.

---

## 3) Toolchain & Install

**Homebrew packages**
```bash
# Core
brew install go golangci-lint git pre-commit make
````

**Go tools (install with Go)**

```bash
# OpenAPI -> Go codegen
go install github.com/ogen-go/ogen/cmd/ogen@latest

# Mock generator (Uber's)
go install go.uber.org/mock/mockgen@latest
```

**Editor**

* Use Go tooling with LSP: `gopls`, `gofmt`, `goimports`.
* Tabs, UTF‑8 (see `.editorconfig` in repo root, referenced—not edited here).

---

## 4) Repository Conventions (assumed)

* **Interfaces** in `internal/<domain>`; **implementations** in `internal/<domain>/v1` (or `pkg/<domain>/v1` if public).
* **Main binaries** under `cmd/<app>/main.go`.
* **OpenAPI source** at `api/openapi.yaml` (or `.json`).
* **ogen output** kept internal, e.g., `internal/api/v1` (do **not** export generated code unless needed).
* **HTTP handlers** implement the ogen‑generated `Server` interface in `internal/controllers/<feature>`.

> Architectural specifics live in `docs/ai/context/systemDesign.md`.

---

## 5) OpenAPI‑First with `ogen`

**Source of truth**

* `api/openapi.yaml` defines HTTP API (OpenAPI v3.x).

**Code generation (example)**

```bash
# Generate server/client/types into an internal package
go run github.com/ogen-go/ogen/cmd/ogen -convenient-errors --clean --package api --target internal/api/v1 api/openapi.yml
```

**Typical wiring (descriptive)**

* Generated interfaces/types: `internal/api/v1/...`
* Implement server methods in `internal/controller/...` where handlers satisfy the generated `Server` interface.
* Compose middlewares (auth/logging/metrics) in `internal/http`.

> Agents: only document or adjust **the command and target paths** here. Do **not** modify code outside `docs/ai/`.

---

## 6) Build, Lint, Test (standard commands)

> These describe expected behavior. Actual wiring may live in Make/Task—**do not** change it here.

* **Format**

  ```bash
  go fmt ./...
  ```

* **Lint**

  ```bash
  golangci-lint run
  ```

* **Unit tests**

  ```bash
  go test ./... -race -count=1
  ```

* **Coverage**

  ```bash
  go test ./... -race -coverprofile=coverage.out
  go tool cover -func=coverage.out
  ```

* **Build (example app)**

  ```bash
  go build -o bin/<app> ./cmd/<app>
  ```

---

## 7) Security & Compliance (dev‑time expectations)

* No secrets in code or `docs/ai/`. Use env vars or secret managers.
* Prefer least‑privilege local credentials.
* Redact sensitive logs (tokens, passwords, PII).
* Third‑party code: prefer well‑maintained libs; avoid unsafe/CGO unless justified.

Repo‑specific security policies belong in `docs/ai/policies/security-architecture.md`.

---

## 8) Agent Usage Notes

* Agents must **consult `docs/ai/` first**.
* Role prompts live under `docs/ai/prompts/`.
* **Foundations** enforces "`docs/ai` only".

# Development Environment

This document outlines the canonical development environment setup and tooling used by all contributors.

---

## Target Platform

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

## Required Versions

- **Go:** `>= 1.22` (use latest stable patch)
- **golangci-lint:** pinned in CI; local can follow latest
- **ogen:** latest stable (OpenAPI‑first code generation)
- **Mock generator (optional):** **Uber** mockgen (`go.uber.org/mock/mockgen`)

> If your local setup needs different versions, document them in your local development notes.

---

## Toolchain Installation

### Homebrew packages
```bash
# Core development tools
brew install go golangci-lint git pre-commit make
```

### Go tools (install with Go)

```bash
# OpenAPI -> Go codegen
go install github.com/ogen-go/ogen/cmd/ogen@latest

# Mock generator (Uber's)
go install go.uber.org/mock/mockgen@latest
```

### Editor Setup

* Use Go tooling with LSP: `gopls`, `gofmt`, `goimports`.
* Configure for tabs, UTF‑8 (see `.editorconfig` in repo root).
* Recommended extensions: Go language support, golangci-lint integration

---

## Repository Conventions

* **Interfaces** in `internal/<domain>`; **implementations** in `internal/<domain>/v1` (or `pkg/<domain>/v1` if public).
* **Main binaries** under `cmd/<app>/main.go`.
* **OpenAPI source** at `api/openapi.yml` (or `.json`).
* **ogen output** kept internal, e.g., `internal/api/v1` (do **not** export generated code unless needed).
* **HTTP handlers** implement the ogen‑generated `Server` interface in `internal/controllers/<feature>`.

> Architectural specifics are documented in `docs/architecture/`.

---

## OpenAPI‑First Development with ogen

### Source of truth

* `api/openapi.yml` defines HTTP API (OpenAPI v3.x).

### Code generation

```bash
# Generate server/client/types into an internal package
go run github.com/ogen-go/ogen/cmd/ogen -convenient-errors --clean --package api --target internal/api/v1 api/openapi.yml
```

### Typical integration

* Generated interfaces/types: `internal/api/v1/...`
* Implement server methods in `internal/controller/...` where handlers satisfy the generated `Server` interface.
* Compose middlewares (auth/logging/metrics) in `internal/http`.

---

## Build, Lint, Test Commands

### Format
```bash
go fmt ./...
```

### Lint
```bash
golangci-lint run
```

### Unit tests
```bash
go test ./... -race -count=1
```

### Coverage
```bash
go test ./... -race -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Build
```bash
go build -o bin/<app> ./cmd/<app>
```

### Development cycle (via Makefile)
```bash
make generate fmt lint test
```

See `CLAUDE.md` for the complete list of available make targets.

---

## Security & Compliance

* No secrets in code or documentation. Use env vars or secret managers.
* Prefer least‑privilege local credentials.
* Redact sensitive logs (tokens, passwords, PII).
* Third‑party code: prefer well‑maintained libs; avoid unsafe/CGO unless justified.

Project-specific security policies are documented in `docs/architecture/` and `docs/guidelines/`.

---

## Troubleshooting

### Common Issues

**Go module issues:**
```bash
go mod tidy
go mod vendor  # if using vendor/
```

**golangci-lint failures:**
```bash
golangci-lint run --fix  # auto-fix where possible
```

**OpenAPI generation issues:**
```bash
make generate  # regenerate from OpenAPI spec
```

### Environment Validation

```bash
make env  # validate development environment setup
```

---

## IDE Integration

This project works well with:
- **VS Code** with Go extension
- **GoLand** from JetBrains
- **Vim/Neovim** with Go LSP setup
- Any editor with `gopls` language server support

Configure your editor to:
- Use `gopls` for Go language support
- Run `gofmt`/`goimports` on save
- Integrate with `golangci-lint`
- Support Go debugging

---

*For core development principles and workflows, see `docs/development/foundations.md`. For project-specific context, see `docs/development/project-context.md`.*
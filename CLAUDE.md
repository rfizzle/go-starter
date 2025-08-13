# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Essential Commands
- **Build**: `make build` - Build all binaries for multiple platforms (Linux, macOS, Windows)
- **Test**: `make test` - Run all Go tests in the project
- **Lint**: `make lint` - Run go vet and golangci-lint for code quality checks
- **Format**: `make fmt` - Format Go code with gofmt, goimports, and YAML files with yamlfmt
- **Generate**: `make generate` - Run go generate to regenerate API code from OpenAPI spec
- **Development cycle**: `make generate fmt lint test` - Full development workflow

### Additional Commands  
- **Single platform builds**: `make linux`, `make osx`, `make windows`
- **Dependencies**: `make tidy` (rebuild go.mod), `make vendor` (sync to vendor/)
- **Environment check**: `make env` - Validate development environment setup
- **Race condition testing**: `make race` - Run tests with race detector
- **CI pipeline**: `make ci` - Run tests and linting suitable for CI environments
- **Docker**: `make docker` - Build Docker container using build scripts

### NPM Scripts (Alternative)
The project also supports npm-style commands via package.json:
- `npm run build`, `npm run test`, `npm run fmt`, `npm run lint`, `npm run tidy`, `npm run vendor`

## Project Architecture

### Code Generation Strategy
This project uses **ogen** for OpenAPI-first API development:
- OpenAPI spec: `api/openapi.yml` - Single source of truth for API definition
- Generated code: `internal/api/v1/oas_*_gen.go` - Auto-generated from OpenAPI spec
- Generation trigger: `go generate` or `make generate`
- **Never edit generated files directly** - all API changes must be made in the OpenAPI spec

### Application Structure
- **Entry point**: `main.go` → `cmd/serve.go` - CLI-based application using Cobra
- **Application core**: `internal/app/app.go` - Main application lifecycle management
- **Web server**: `internal/web/webserver.go` - Chi-based HTTP server with middleware stack
- **Controllers**: `internal/controller/` - Business logic handlers that implement generated interfaces
- **Storage interface**: `internal/storage/driver` - Database abstraction driver
- **Storage layer**: `internal/storage/postgres` - PostgreSQL implementation of the database driver abstraction
- **Authentication**: `internal/auth/` - JWT-based auth system with middleware

### Key Dependencies
- **Web framework**: Chi v5 for HTTP routing and middleware
- **Database**: PostgreSQL with pgx/v5 driver and scany for scanning
- **API generation**: ogen for OpenAPI code generation
- **Configuration**: Viper for config management, Cobra for CLI
- **Logging**: Zap for structured logging
- **Observability**: Prometheus metrics, OpenTelemetry tracing, pprof debugging

### Development Workflow
1. **API changes**: Modify `api/openapi.yml` first
2. **Generate code**: Run `make generate` to update generated API files
3. **Implement handlers**: Update controllers in `internal/controller/` to implement new interfaces
4. **Format and lint**: Run `make fmt lint` before committing
5. **Test**: Run `make test` to ensure everything works
6. **Build**: Use `make build` for multi-platform binaries

### Middleware Stack
The web server includes these middleware layers (in order):
- Timeout (60s), RequestID, RealIP, Recoverer
- Custom logging middleware for API routes
- Static file serving
- CORS configuration
- Built-in routes: `/metrics` (Prometheus), `/debug/pprof/*` (profiling)

### Configuration
- CLI flags and environment variables supported via Viper
- Configuration structure defined in `pkg/config/config.go`
- Webserver settings: address, port, timeouts configurable via flags or env vars
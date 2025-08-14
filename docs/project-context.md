# Project Context

## Overview
Go Starter is a production-ready, enterprise-grade foundation for building scalable REST API services. It provides OpenAPI-first development with automatic code generation, JWT authentication, PostgreSQL integration, and comprehensive observability.

## Key Goals
- Speed up API development with an 80% time reduction
- Maintain >90% test coverage
- Zero additional configs for Kubernetes deployment

## Architecture
- **Entry point**: `main.go` → `cmd/serve.go` (Cobra CLI)
- **Application core**: `internal/app/app.go`
- **Web server**: `internal/web/webserver.go` (Chi-based)
- **Controllers**: `internal/controller/` (implement generated interfaces)
- **Storage**: `internal/storage/driver` (abstraction) + `internal/storage/postgres`
- **Authentication**: `internal/auth/` (JWT + middleware)

## Tech Stack
- **Language**: Go (latest stable)
- **Web**: Chi router + CORS middleware
- **API**: OpenAPI 3.x with ogen code generation
- **Database**: PostgreSQL + pgx driver + scany
- **Config**: Viper + Cobra
- **Observability**: Zap logging + Prometheus + OpenTelemetry + pprof
- **Testing**: testify + pgxmock

## Code Generation Strategy
- **Single source of truth**: `api/openapi.yml`
- **Generated code**: `internal/api/v1/oas_*_gen.go`
- **Trigger**: `make generate`
- **Never edit generated files directly**

## Development Workflow
1. Modify `api/openapi.yml`
2. Run `make generate`
3. Update controllers in `internal/controller/`
4. Run `make fmt lint test`
5. Build with `make build`

## Coding Standards
- **Structure**: Follow Go Standard Project Layout
- **Naming**: camelCase variables, PascalCase exports, no snake_case
- **Packages**: Single purpose, short names, avoid util/helpers
- **Dependencies**: Dependency injection with interfaces
- **Errors**: Handle all errors, wrap with context
- **Testing**: Table-driven tests, mock external services
- **Format**: Use goimports, avoid long lines (120 chars)

## Key Commands
- `make build` - Multi-platform binaries
- `make test` - All Go tests
- `make lint` - Code quality checks
- `make fmt` - Format code
- `make generate` - Regenerate API code
- `make ci` - CI pipeline checks
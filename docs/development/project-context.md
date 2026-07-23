# Project Context

## Overview
Go Starter is a production-ready, enterprise-grade foundation for building scalable REST API services. It provides OpenAPI-first development with automatic code generation, JWT authentication, PostgreSQL integration, and comprehensive observability.

## Key Goals
- Speed up API development with an 80% time reduction
- Protect public contracts and significant risks with focused tests
- Zero additional configs for Kubernetes deployment

**Reference `foundations.md` first** - it contains core Go principles, patterns, and workflows that apply to all work.

---

## Development Principles

* **Architecture**: Follow **Go Standard Project Layout** with explicit package contracts
* **Workflow**: Map acceptance criteria and significant risks to effective verification
* **Quality**: Use **`go test`** and **`golangci-lint`** for validation
* **Git**: Clean commits with meaningful messages, feature branches for major changes
* **Clarity**: Clear names, document **why** decisions were made, not just what code does
* **Performance**: Optimize for readability first, performance second (unless performance is key requirement)

---

## Core Workflows

### Planning and Design

When implementing new features or significant changes:

1. **Understand Requirements** - Review specifications and acceptance criteria
2. **Research Existing Code** - Study current patterns and interfaces
3. **Design Approach** - Plan the implementation strategy
4. **Create Implementation Plan** - Break down into manageable tasks
5. **Implement with TDD** - Write tests first, then implementation
6. **Document Changes** - Update relevant documentation

### Development Workflow

1. **API Changes** (if needed):
   - Modify `api/openapi.yml` first
   - Run `make generate` to update generated API files
   - Update controllers in `internal/controller/` to implement new interfaces

2. **Implementation** (contract-driven approach):
   - Create feature branch from main
   - Map acceptance criteria and significant risks to effective verification
   - Implement the smallest change that satisfies the approved contract
   - Refactor while keeping tests green

3. **Quality Checks**:
   - Run `make fmt` to format code
   - Run `make lint` for code quality
   - Run `make test` to ensure all tests pass
   - Run `make build` for multi-platform binaries

4. **Commit & Review**:
   - Atomic commits with clear messages
   - Create pull request for code review

### Key Commands

- `make build` - Build multi-platform binaries
- `make test` - Run all Go tests
- `make lint` - Run code quality checks
- `make fmt` - Format code with gofmt and goimports
- `make generate` - Regenerate API code from OpenAPI spec
- `make ci` - Run CI pipeline checks
- `make race` - Run tests with race detector
- `make docker` - Build Docker container

---

## Documentation Update Guidelines

Update project documentation when:

1. **Discovering new patterns** or best practices specific to this project
2. **Completing significant changes** that affect architecture or workflows
3. **Adding new tools** or changing development environment
4. **Making architectural decisions** that need to be recorded
5. **Creating project-specific conventions** beyond standard Go practices

---

## Architecture

### Application Structure
- **Entry point**: `main.go` → `cmd/serve.go` (Cobra CLI)
- **Composition root**: `internal/app/` - Dependency assembly and application lifecycle
- **Domain contracts**: `internal/domain/` - Infrastructure-free business data, interfaces, and errors
- **Business modules**: `internal/modules/` - Business capabilities and their adapters
- **Platform**: `internal/platform/` - Configuration, database, middleware, observability, rate limiting, and web infrastructure
- **Controllers**: `internal/controller/` - Generated transport-to-domain mapping
- **Authentication**: `internal/auth/` - JWT validation and generated security integration

### Key Patterns
* **Module/Platform Boundaries** - Modules may depend on domain contracts and selected platform capabilities; domain and platform do not depend on modules
* **Consumer-Owned Interfaces** - Business orchestration and edges accept the narrow behavior they consume
* **OpenAPI-First Development** - API definitions drive code generation
* **Explicit Composition** - Concrete dependencies are assembled in `internal/app/` or a CLI composition root

See [the accepted boundary decision](../architecture/module-platform-boundaries.md).

---

## Tech Stack

- **Language**: Go 1.22+ (latest stable)
- **Web Framework**: Chi v5 for HTTP routing and middleware
- **API Generation**: ogen for OpenAPI code generation
- **Database**: PostgreSQL with pgx/v5, sqlc, squirrel, and Goose
- **Configuration**: Viper for config management, Cobra for CLI
- **Logging**: Zap for structured logging
- **Observability**: Prometheus metrics, OpenTelemetry tracing, pprof debugging
- **Testing**: testify assertions, pgxmock for database mocking
- **Code Quality**: golangci-lint, gofmt, goimports

## Code Generation Strategy

- **Single source of truth**: `api/openapi.yml`
- **Generated code**: `internal/api/v1/oas_*_gen.go`
- **Generation trigger**: `make generate` or `go generate`
- **Never edit generated files directly** - all API changes must be made in the OpenAPI spec

---

## Quality Standards

All code must meet these standards before merging:

* **Tests Pass** - All existing and new tests must pass
* **Useful Tests** - Each test protects a contract, regression, boundary, or named risk
* **Linting Clean** - No golangci-lint warnings
* **Formatted** - Code properly formatted with gofmt and goimports
* **Documentation** - Public APIs documented with godoc comments
* **Naming Conventions** - camelCase variables, PascalCase exports, no snake_case
* **Error Handling** - All errors handled and wrapped with context
* **Integration** - Works correctly with existing codebase

## Middleware Stack

The web server includes these middleware layers (in order):
- Timeout (60s default)
- RequestID generation
- RealIP extraction
- Panic recovery
- Custom logging middleware for API routes
- Static file serving
- CORS configuration
- Built-in routes: `/metrics` (Prometheus), `/debug/pprof/*` (profiling)

---

*For core Go principles and development patterns, see `foundations.md`. For specific guidelines and checklists, see the `docs/guidelines/` directory.*

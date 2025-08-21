# Architecture Documentation

This directory contains architectural decisions, patterns, and system design documentation for the Go Starter project.

## Purpose

Document and track architectural decisions that shape the project's structure, patterns, and technical choices. This ensures consistency, provides historical context, and guides future development.

## Contents

### Current Documents
- *Documentation is being migrated and organized*

### Planned Documents
- **current-state.md** - Current implementation status and architecture overview
- **adr-001-openapi-first.md** - Decision to use OpenAPI-first development
- **adr-002-interface-pattern.md** - Interface-implementation package pattern
- **adr-003-ogen-codegen.md** - Choice of ogen for code generation
- **adr-004-database-abstraction.md** - Database driver abstraction pattern

## Architecture Overview

### System Architecture
The project follows a layered architecture:
1. **API Layer** - OpenAPI specification and generated code
2. **Controller Layer** - Business logic handlers
3. **Service Layer** - Core business logic (interfaces and implementations)
4. **Repository Layer** - Data persistence abstraction
5. **Infrastructure Layer** - Database, external services

### Key Patterns

#### OpenAPI-First Development
- Single source of truth: `api/openapi.yml`
- Generated code in `internal/api/v1/`
- Controllers implement generated interfaces

#### Interface-Implementation Pattern
- Interfaces at package level: `internal/<domain>/`
- Implementations in versioned subdirs: `internal/<domain>/v1/`
- Enables multiple implementations and versioning

#### Dependency Injection
- Constructor functions accept interfaces
- Dependencies wired in application layer
- Facilitates testing and flexibility

## Architecture Decision Records (ADRs)

ADRs document significant architectural decisions. Each ADR should include:

1. **Title** - Short descriptive title
2. **Status** - Proposed, Accepted, Deprecated, Superseded
3. **Context** - What prompted this decision
4. **Decision** - What was decided
5. **Consequences** - Both positive and negative outcomes
6. **Alternatives** - Other options considered

## Guidelines for Contributors

When documenting architecture:
1. Use ADR format for significant decisions
2. Include diagrams where helpful
3. Link to relevant code examples
4. Update current-state.md when implementing changes
5. Consider impact on existing patterns

## Related Documentation

- [../development/project-context.md](../development/project-context.md) - Project overview and structure
- [../guidelines/interface-pattern.md](../guidelines/interface-pattern.md) - Detailed interface pattern guide
- [../development/foundations.md](../development/foundations.md) - Core Go principles

## Tools and Technologies

### Core Stack
- **Language**: Go 1.22+
- **Web Framework**: Chi v5
- **API Generation**: ogen (OpenAPI generator)
- **Database**: PostgreSQL with pgx driver
- **Configuration**: Viper + Cobra

### Supporting Tools
- **Logging**: Zap
- **Metrics**: Prometheus
- **Tracing**: OpenTelemetry
- **Testing**: testify, pgxmock

---

*This directory helps maintain architectural consistency and provides context for technical decisions throughout the project's evolution.*
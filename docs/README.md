# Documentation Index

Welcome to the Go Starter project documentation. This index provides a comprehensive overview of all documentation and helps both humans and AI agents navigate the codebase effectively.

## Quick Start

1. **New to the project?** Start with [development/project-context.md](development/project-context.md)
2. **Setting up development?** See [development/environment.md](development/environment.md)
3. **Understanding Go patterns?** Read [development/foundations.md](development/foundations.md)
4. **Working with AI agents?** Check [ai/README.md](ai/README.md)

## Documentation Structure

### 📁 development/
Core development documentation for engineers working on this codebase.

- **[project-context.md](development/project-context.md)** - Project overview, architecture, tech stack, and workflows
- **[foundations.md](development/foundations.md)** - Core Go principles, patterns, and best practices
- **[environment.md](development/environment.md)** - Development environment setup, tools, and troubleshooting
- **[database.md](development/database.md)** - PostgreSQL, sqlc, migrations, repositories, and integration testing
- **[ogen-workflow.md](development/ogen-workflow.md)** - OpenAPI-first generation and controller workflow

### 📁 guidelines/
Specific development guidelines and standards for consistent code quality.

- **[go-standards.md](guidelines/go-standards.md)** - Go-specific coding standards and conventions
- **[interface-pattern.md](guidelines/interface-pattern.md)** - Consumer-owned interface and implementation guidance
- **[testing-strategy.md](guidelines/testing-strategy.md)** - Risk-based testing and contract verification

### 📁 architecture/
Architectural decisions, patterns, and system design documentation.

- **[module-platform-boundaries.md](architecture/module-platform-boundaries.md)** - Accepted package placement, dependency direction, and enforcement
- **[current-state.md](architecture/current-state.md)** - Implemented runtime and package map

### 📁 features/
Feature specifications and requirements documentation.

- Feature documents follow the template in [ai/templates/feature.md](ai/templates/feature.md)
- Each feature includes user stories, acceptance criteria, and technical requirements

### 📁 plans/
Implementation plans created from feature specifications.

- Plans follow the template in [ai/templates/plan.md](ai/templates/plan.md)
- Include task breakdowns, technical approaches, and risk assessments

### 📁 ai/
AI agent documentation, prompts, and templates for consistent AI interactions.

#### ai/prompts/
AI agent role definitions and scope boundaries.

- **[documentation.md](ai/prompts/documentation.md)** - Documentation and feature-specification agent
- **[planning.md](ai/prompts/planning.md)** - Senior engineer agent for implementation planning
- **[review.md](ai/prompts/review.md)** - Code review agent for quality assurance

#### ai/templates/
Reusable templates for creating consistent documentation.

- **[README.md](ai/templates/README.md)** - Template usage guide
- **[feature.md](ai/templates/feature.md)** - Feature specification template
- **[plan.md](ai/templates/plan.md)** - Implementation plan template
- **[prompt.md](ai/templates/prompt.md)** - AI agent prompt template

#### ai/
- **[README.md](ai/README.md)** - AI system overview and agent coordination

## Key Concepts

### OpenAPI-First Development
This project follows an OpenAPI-first approach where:
1. API definitions in `api/openapi.yml` drive code generation
2. Controllers implement generated interfaces
3. Changes start with the OpenAPI specification

### Package Contracts and Implementations
Every nontrivial package has a documented contract. An interface is added only for a real
substitution boundary or a consumer that needs a narrower seam:
- Keep one natural implementation in the contract package
- Keep multiple real implementations in named subpackages
- Prefer small consumer-owned interfaces and concrete constructor returns

### Module and Platform Layout
Application code is divided into `internal/modules` for business capabilities,
`internal/platform` for technical infrastructure, and `internal/domain` for
shared infrastructure-free contracts. The composition root lives in
`internal/app`; generated transport remains quarantined under `internal/api`.

## Development Workflow

### Standard Development Cycle
1. **API Changes** (if needed): Modify `api/openapi.yml`
2. **Generate Code**: Run `make generate`
3. **Select Verification**: Map requirements and risks to the narrowest effective checks
4. **Implement**: Make the smallest change that satisfies the approved contract
5. **Quality Checks**: Run focused tests, then `make check`
6. **Build**: Run `make build`

### Key Commands
- `make build` - Build the local binary in `bin/`
- `make test` - Run tests that do not require external services
- `make test-integration` - Run disposable-service integration tests
- `make test-all` - Run both test tiers
- `make lint` - Code quality checks
- `make fmt` - Format code
- `make generate` - Regenerate API code
- `make docs-check` - Validate documentation links, commands, paths, and Go version
- `make ci` - CI pipeline checks

## For AI Agents

### Navigation Tips
1. Start with [ai/README.md](ai/README.md) for AI system overview
2. Use [development/project-context.md](development/project-context.md) for project understanding
3. Follow templates in [ai/templates/](ai/templates/) for consistent documentation
4. Check agent prompts in [ai/prompts/](ai/prompts/) for scope definitions

### Best Practices
- Always reference specific file paths when discussing code
- Use templates for creating new documentation
- Define clear package contracts and add interfaces only at justified seams
- Maintain consistency with existing patterns

## Contributing

When adding new documentation:
1. Place files in the appropriate directory
2. Update this README.md index
3. Follow existing naming conventions
4. Use templates where applicable
5. Cross-reference related documents

## Version Information

- **Go Version**: 1.26.5 (from `go.mod`)
- **Project Type**: REST API service
- **Architecture**: OpenAPI-first with ogen code generation
- **Database**: PostgreSQL with pgx driver
- **Web Framework**: Chi router v5

---

*This documentation structure enables effective collaboration between human developers and AI agents while maintaining high code quality and consistency.*

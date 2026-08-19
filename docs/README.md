# Documentation Index

Welcome to the Go Starter project documentation. This index provides a comprehensive overview of all documentation and helps both humans and AI agents navigate the codebase effectively.

## Quick Start

1. **New to the project?** Start with [development/project-context.md](development/project-context.md)
2. **Setting up development?** See [development/environment.md](development/environment.md)
3. **Understanding Go patterns?** Read [development/foundations.md](development/foundations.md)
4. **Working with AI agents?** Read [AGENTS.md](../AGENTS.md), then [ai/README.md](ai/README.md)

## Documentation Structure

### 📁 development/
Core development documentation for engineers working on this codebase.

- **[README.md](development/README.md)** - Development documentation overview
- **[project-context.md](development/project-context.md)** - Project overview, architecture, tech stack, and workflows
- **[foundations.md](development/foundations.md)** - Core Go principles, patterns, and best practices
- **[environment.md](development/environment.md)** - Development environment setup, tools, and troubleshooting
- **[database.md](development/database.md)** - PostgreSQL, sqlc, migrations, repositories, and integration testing
- **[ogen-workflow.md](development/ogen-workflow.md)** - OpenAPI-first generation and controller workflow
- **[ci-contract.md](development/ci-contract.md)** - What each GitHub Actions workflow enforces, release provenance, and tool-pin maintenance

### 📁 guidelines/
Specific development guidelines and standards for consistent code quality.

- **[README.md](guidelines/README.md)** - Guideline index and scope
- **[go-standards.md](guidelines/go-standards.md)** - Go-specific coding standards and conventions
- **[interface-pattern.md](guidelines/interface-pattern.md)** - Consumer-owned interface and implementation guidance
- **[testing-strategy.md](guidelines/testing-strategy.md)** - Risk-based testing and contract verification

### 📁 architecture/
Architectural decisions, patterns, and system design documentation.

- **[README.md](architecture/README.md)** - Architecture documentation overview
- **[module-platform-boundaries.md](architecture/module-platform-boundaries.md)** - Accepted package placement, dependency direction, and enforcement
- **[current-state.md](architecture/current-state.md)** - Implemented runtime and package map

### 📁 reviews/
Point-in-time code review records. These are historical artifacts, not current
guidance — they are excluded from the documentation linter's freshness checks.

- **[README.md](reviews/README.md)** - Review workflow and document conventions
- **[2026-07-03-starter-hardening-review.md](reviews/2026-07-03-starter-hardening-review.md)** - Starter hardening review

### 📁 ai/
Supporting material for AI-assisted work. [AGENTS.md](../AGENTS.md) holds the
authoritative rules and the task type → skill → verification routing table; the
skill library itself lives in [`.agents/skills/`](../.agents/skills/README.md).

- **[README.md](ai/README.md)** - How AGENTS.md, skills, role agents, and templates fit together
- **[agent-guide.md](ai/agent-guide.md)** - Role agent scopes and handoff points
- **[workflow-shortcuts.md](ai/workflow-shortcuts.md)** - Common multi-step task patterns
- **[DEFINITIONS.md](ai/DEFINITIONS.md)** - Shared terminology and quality gates

#### ai/prompts/
Role agent definitions, surfaced to Claude Code as relative symlinks under `.claude/agents/`.

- **[documentation.md](ai/prompts/documentation.md)** - Documentation and feature-specification agent
- **[planning.md](ai/prompts/planning.md)** - Senior engineer agent for implementation planning
- **[review.md](ai/prompts/review.md)** - Code review agent for quality assurance
- **[development.md](ai/prompts/development.md)** - Implementation and bug-fix agent

#### ai/templates/
Reusable templates for creating consistent documentation.

- **[README.md](ai/templates/README.md)** - Template usage guide
- **[feature.md](ai/templates/feature.md)** - Feature specification template
- **[plan.md](ai/templates/plan.md)** - Implementation plan template
- **[review.md](ai/templates/review.md)** - Code review report template
- **[prompt.md](ai/templates/prompt.md)** - AI agent prompt template

## Key Concepts

### OpenAPI-First Development
This project follows an OpenAPI-first approach where:
1. API definitions in `api/openapi.yml` drive code generation
2. Controllers implement generated interfaces
3. Changes start with the OpenAPI specification

### Package Contracts and Implementations
Every nontrivial package documents its contract in a package doc comment; generated
packages (`internal/api/v1`, `internal/platform/database/postgres/db`) are exempt because
their sources define the contract. An interface is added only for a real substitution
boundary or a consumer that needs a narrower seam:
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
- `make secrets-check` - Scan the git history for committed secrets
- `make ci` - CI pipeline checks (see [development/ci-contract.md](development/ci-contract.md))

## For AI Agents

### Navigation Tips
1. Start with [AGENTS.md](../AGENTS.md) — it is authoritative and carries the task type → skill → verification routing table
2. Load the narrowest applicable skill from [`.agents/skills/`](../.agents/skills/README.md)
3. Use [development/project-context.md](development/project-context.md) for project understanding
4. See [ai/README.md](ai/README.md) for how the role agents and templates fit around AGENTS.md
5. Follow templates in [ai/templates/](ai/templates/) for consistent documentation

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

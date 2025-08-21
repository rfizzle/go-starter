# Documentation Index

Welcome to the Go Starter project documentation. This index provides a comprehensive overview of all documentation and helps both humans and AI agents navigate the codebase effectively.

## Quick Start

1. **New to the project?** Start with [development/project-context.md](development/project-context.md)
2. **Setting up development?** See [development/environment.md](development/environment.md)
3. **Understanding Go patterns?** Read [development/foundations.md](development/foundations.md)
4. **Working with AI agents?** Check [ai/REFERENCES.md](ai/REFERENCES.md)

## Documentation Structure

### 📁 development/
Core development documentation for engineers working on this codebase.

- **[project-context.md](development/project-context.md)** - Project overview, architecture, tech stack, and workflows
- **[foundations.md](development/foundations.md)** - Core Go principles, patterns, and best practices
- **[environment.md](development/environment.md)** - Development environment setup, tools, and troubleshooting

### 📁 guidelines/
Specific development guidelines and standards for consistent code quality.

- **[go-standards.md](guidelines/go-standards.md)** - Go-specific coding standards and conventions
- **[interface-pattern.md](guidelines/interface-pattern.md)** - Interface-implementation pattern guide with examples
- **[testing-strategy.md](guidelines/testing-strategy.md)** - Comprehensive testing approach and requirements

### 📁 architecture/
Architectural decisions, patterns, and system design documentation.

- *Currently being populated with architectural decision records (ADRs)*
- Future: `current-state.md` - Current implementation status

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

- **[ai.md](ai/prompts/ai.md)** - AI documentation specialist agent
- **[feature.md](ai/prompts/feature.md)** - Product owner agent for feature requirements
- **[plan.md](ai/prompts/plan.md)** - Senior engineer agent for implementation planning
- **[review.md](ai/prompts/review.md)** - Code review agent for quality assurance

#### ai/templates/
Reusable templates for creating consistent documentation.

- **[README.md](ai/templates/README.md)** - Template usage guide
- **[feature.md](ai/templates/feature.md)** - Feature specification template
- **[plan.md](ai/templates/plan.md)** - Implementation plan template
- **[prompt.md](ai/templates/prompt.md)** - AI agent prompt template

#### ai/
- **[REFERENCES.md](ai/REFERENCES.md)** - File reference guide for AI agents

## Key Concepts

### OpenAPI-First Development
This project follows an OpenAPI-first approach where:
1. API definitions in `api/openapi.yml` drive code generation
2. Controllers implement generated interfaces
3. Changes start with the OpenAPI specification

### Interface-Implementation Pattern
All domain packages follow a consistent pattern:
- Interfaces defined at package level (`internal/<domain>/`)
- Implementations in versioned subdirectories (`internal/<domain>/v1/`)
- See [guidelines/interface-pattern.md](guidelines/interface-pattern.md) for details

### Standard Go Project Layout
The codebase follows the Standard Go Project Layout:
- `/cmd` - Main applications
- `/internal` - Private application code
- `/pkg` - Public libraries
- `/api` - OpenAPI specifications
- `/scripts` - Build and utility scripts
- `/docs` - Documentation (this directory)

## Development Workflow

### Standard Development Cycle
1. **API Changes** (if needed): Modify `api/openapi.yml`
2. **Generate Code**: Run `make generate`
3. **Write Tests**: Follow TDD approach
4. **Implement**: Write code to pass tests
5. **Quality Checks**: Run `make fmt lint test`
6. **Build**: Run `make build`

### Key Commands
- `make build` - Build multi-platform binaries
- `make test` - Run all tests
- `make lint` - Code quality checks
- `make fmt` - Format code
- `make generate` - Regenerate API code
- `make ci` - CI pipeline checks

## For AI Agents

### Navigation Tips
1. Start with [ai/REFERENCES.md](ai/REFERENCES.md) for file path references
2. Use [development/project-context.md](development/project-context.md) for project understanding
3. Follow templates in [ai/templates/](ai/templates/) for consistent documentation
4. Check agent prompts in [ai/prompts/](ai/prompts/) for scope definitions

### Best Practices
- Always reference specific file paths when discussing code
- Use templates for creating new documentation
- Follow the interface-implementation pattern for new code
- Maintain consistency with existing patterns

## Contributing

When adding new documentation:
1. Place files in the appropriate directory
2. Update this README.md index
3. Follow existing naming conventions
4. Use templates where applicable
5. Cross-reference related documents

## Version Information

- **Go Version**: 1.22+ required
- **Project Type**: REST API service
- **Architecture**: OpenAPI-first with ogen code generation
- **Database**: PostgreSQL with pgx driver
- **Web Framework**: Chi router v5

---

*This documentation structure enables effective collaboration between human developers and AI agents while maintaining high code quality and consistency.*
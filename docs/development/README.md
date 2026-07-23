# Development Documentation

Core development documentation for engineers working on the Go Starter project. This directory contains essential information about project context, Go patterns, and development environment setup.

## Quick Reference

### Essential Reading Order
1. **[project-context.md](project-context.md)** - Start here for project overview and workflows
2. **[foundations.md](foundations.md)** - Core Go principles and patterns (must read)
3. **[environment.md](environment.md)** - Development environment setup and troubleshooting

### Key Commands
```bash
make generate    # Regenerate API code from OpenAPI spec
make architecture-check # Verify package placement and import direction
make fmt         # Format all Go and YAML files
make lint        # Run code quality checks
make test        # Run all tests
make build       # Build multi-platform binaries
make ci          # Full CI pipeline (generate, fmt, lint, test)
```

## Documentation Overview

### [project-context.md](project-context.md)
**Purpose**: Project-specific context, architecture, and workflows  
**Audience**: All developers (human and AI)  
**Contains**:
- Project overview and key goals
- Architecture and tech stack
- Development workflows and commands
- Quality standards and guidelines

### [foundations.md](foundations.md)
**Purpose**: Core Go development principles and patterns  
**Audience**: All Go developers  
**Contains**:
- Go standards and conventions
- Interface design patterns
- Error handling approaches
- Code quality guidelines

### [environment.md](environment.md)
**Purpose**: Development environment setup and tooling  
**Audience**: New developers and CI/CD systems  
**Contains**:
- Go version requirements
- Development tools setup
- Build and test configurations
- Troubleshooting guides

## Development Workflows

### Standard Development Cycle

1. **Plan**: Review requirements, package placement, and the applicable module spec
2. **API First**: Update `api/openapi.yml` if API changes are needed
3. **Generate**: Run `make generate` to update generated code
4. **Verify**: Select tests for acceptance criteria and significant risks
5. **Implement**: Make the smallest change that satisfies the approved contract
6. **Quality**: Run `make architecture-check`, then `make fmt lint test`
7. **Review**: Submit for code review
8. **Deploy**: Build and deploy

### Common Tasks

#### Adding an API Endpoint
1. Update `api/openapi.yml` with endpoint definition
2. Run `make generate` to create interfaces
3. Implement controller in `internal/controller/`
4. Add tests for the controller
5. Update integration tests

#### Creating a New Service
Reference [docs/guidelines/interface-pattern.md](../guidelines/interface-pattern.md) for
contract-first package and consumer-owned interface guidance.

#### Database Changes
1. Create a sequential migration in `internal/platform/database/postgres/migrations/`
2. Update repository interfaces
3. Implement repository methods
4. Add repository tests
5. Run migration in development

## Quality Standards

All code must meet these standards:

### Go Standards
- **Formatting**: `gofmt` and `goimports` compliant
- **Linting**: Pass `golangci-lint` checks
- **Testing**: Tests protect named contracts, regressions, boundaries, or risks
- **Documentation**: Godoc comments on all exported items
- **Naming**: Follow Go naming conventions

### Project Standards
- **Package Contracts**: Document purpose, non-goals, public behavior, bounds, and ownership
- **Error Handling**: Proper error wrapping with context
- **Context Usage**: Thread context through all operations
- **Test Selection**: Use the narrowest stable test boundary that detects the relevant failure

### Architecture Standards
- **Standard Go Project Layout**: Consistent directory structure
- **OpenAPI First**: API changes start with specification
- **Clean Architecture**: Proper separation of concerns
- **Dependency Injection**: Small consumer-owned interfaces at real seams

## Troubleshooting

### Build Issues
```bash
make clean       # Clean generated files and build artifacts
make tidy        # Update go.mod and dependencies
make vendor      # Sync vendor directory
```

### Generated Code Issues
```bash
# Clean and regenerate API code
make clean-gen && make generate
```

### Test Issues
```bash
make test-race   # Run tests with race detector
make test-cover  # Run tests with coverage report
```

## For New Developers

### Getting Started
1. Read this entire directory (project-context → foundations → environment)
2. Set up development environment following environment.md
3. Run `make ci` to verify everything works
4. Read through the guidelines directory
5. Look at existing code to understand patterns

### First Contribution Checklist
- [ ] Development environment set up and tested
- [ ] All quality checks passing (`make ci`)
- [ ] Familiar with package contracts and consumer-owned interface guidance
- [ ] Understanding of OpenAPI workflow
- [ ] Risk-based testing approach understood and practiced
- [ ] Code review process understood

---

*This development documentation provides the foundation for all work on the Go Starter project. Keep it current as the project evolves.*

# Development Documentation

Core development documentation for engineers working on the Go Starter project. Start here to understand the project, set up your environment, and learn development patterns.

## Quick Start Guide

1. **Read** [project-context.md](project-context.md) - Understand the project
2. **Setup** [environment.md](environment.md) - Configure your development environment  
3. **Learn** [foundations.md](foundations.md) - Master Go patterns and principles

## Contents

### [project-context.md](project-context.md)
**Project overview and development workflows**
- Project goals and architecture
- Tech stack and dependencies
- Development workflow and commands
- Quality standards and conventions

### [foundations.md](foundations.md)
**Core Go principles and patterns**
- Go idioms and best practices
- Error handling patterns
- Testing strategies
- Concurrency patterns
- Performance considerations

### [environment.md](environment.md)
**Development environment setup and tools**
- Go version and toolchain setup
- Required dependencies
- IDE configuration
- Troubleshooting guide

### Planned Additions
- **ogen-workflow.md** - Complete guide for OpenAPI code generation
- **database.md** - Database patterns and migration strategies
- **debugging.md** - Debugging techniques and tools

## Development Workflow

### Daily Development Cycle
```bash
# 1. Start with latest code
git pull origin main

# 2. Create feature branch
git checkout -b feature/your-feature

# 3. Make API changes (if needed)
# Edit api/openapi.yml

# 4. Generate code
make generate

# 5. Write tests (TDD)
# Create *_test.go files

# 6. Implement feature
# Write code to pass tests

# 7. Quality checks
make fmt      # Format code
make lint     # Check code quality
make test     # Run tests

# 8. Build and verify
make build    # Build binaries

# 9. Commit and push
git add .
git commit -m "feat: your feature"
git push origin feature/your-feature
```

### Essential Commands

| Command | Description |
|---------|-------------|
| `make build` | Build multi-platform binaries |
| `make test` | Run all tests |
| `make lint` | Run linting checks |
| `make fmt` | Format code |
| `make generate` | Generate API code |
| `make race` | Test for race conditions |
| `make ci` | Run CI pipeline locally |
| `make docker` | Build Docker image |

## Development Principles

### Test-Driven Development (TDD)
1. Write failing test first
2. Write minimal code to pass
3. Refactor while keeping tests green
4. Aim for >90% test coverage

### Code Quality Standards
- All tests must pass
- No linting warnings
- Properly formatted (gofmt, goimports)
- Clear variable and function names
- Comprehensive error handling
- Godoc comments for public APIs

### Git Workflow
- Feature branches from main
- Atomic, focused commits
- Clear commit messages (conventional commits)
- Pull requests for code review
- Squash and merge to main

## Common Tasks

### Adding a New API Endpoint
1. Update `api/openapi.yml` with endpoint definition
2. Run `make generate` to create interfaces
3. Implement controller in `internal/controller/`
4. Add tests for the controller
5. Update integration tests

### Creating a New Service
1. Define interface in `internal/<domain>/interfaces.go`
2. Create implementation in `internal/<domain>/v1/`
3. Write unit tests for the service
4. Wire up in `internal/app/app.go`
5. Document in architecture docs

### Database Changes
1. Create migration in `migrations/`
2. Update repository interfaces
3. Implement repository methods
4. Add repository tests
5. Run migration in development

## Troubleshooting

### Common Issues

**Build Failures**
- Check Go version: `go version` (need 1.22+)
- Update dependencies: `make tidy`
- Clean and rebuild: `make clean build`

**Test Failures**
- Run specific test: `go test -v -run TestName`
- Check test database: ensure PostgreSQL is running
- Review recent changes: `git diff`

**Generation Issues**
- Validate OpenAPI spec: `make validate`
- Check ogen version: `go list -m github.com/ogen-go/ogen`
- Clean generated files: `rm -rf internal/api/v1/*_gen.go`

## Resources

### Internal Documentation
- [../architecture/](../architecture/) - Architecture decisions
- [../guidelines/](../guidelines/) - Coding standards
- [../ai/](../ai/) - AI agent documentation

### External Resources
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [ogen Documentation](https://ogen.dev)

---

*This directory contains essential documentation for productive development on the Go Starter project.*
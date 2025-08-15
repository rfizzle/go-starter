# Go Development Project Context

This document provides project-specific context and development workflows for engineers working on this Go codebase.

**Reference `foundations.md` first** - it contains core Go principles, patterns, and workflows that apply to all work.

---

## Development Principles

* **Architecture**: Follow **Go Standard Project Layout** with interfaces in their own packages
* **Workflow**: **Test-Driven Development (TDD)** - write tests before code
* **Quality**: Use **`go test`** and **`golangci-lint`** for validation
* **Git**: Clean commits with meaningful messages, feature branches for major changes
* **Clarity**: Clear names, document **why** decisions were made, not just what code does
* **Performance**: Optimize for readability first, performance second (unless performance is key requirement)

---

## Project Documentation Structure

Core engineering documentation in `docs/`:

1. **`development/foundations.md`** - Core Go principles, patterns, workflows (READ FIRST)
2. **`development/project-context.md`** - This file - project-specific context and workflows  
3. **`development/environment.md`** - Go version, dependencies, dev setup, troubleshooting
4. **`architecture/`** - Architectural decisions and patterns
5. **`guidelines/`** - Specific development guidelines and standards

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

### Development Cycle

Standard development workflow:

1. **Create Feature Branch** - Branch from main for new work
2. **Write Tests First** - Follow TDD approach
3. **Implement Code** - Make tests pass with minimal code
4. **Refactor** - Improve code while keeping tests green
5. **Run Quality Checks** - Format, lint, and test
6. **Commit Changes** - Atomic commits with clear messages
7. **Create Pull Request** - Request code review

---

## Documentation Update Guidelines

Update project documentation when:

1. **Discovering new patterns** or best practices specific to this project
2. **Completing significant changes** that affect architecture or workflows
3. **Adding new tools** or changing development environment
4. **Making architectural decisions** that need to be recorded
5. **Creating project-specific conventions** beyond standard Go practices

---

## Project Architecture Notes

This project uses specific patterns documented in `docs/architecture/`:

* **Interface-Implementation Pattern** - Interfaces in domain packages, implementations in versioned subpackages
* **OpenAPI-First Development** - API definitions drive code generation
* **Standard Go Project Layout** - Consistent directory structure and organization

For architectural decisions and patterns specific to this project, see the `docs/architecture/` directory.

---

## Development Environment

For development environment setup, tooling, and troubleshooting, see `docs/development/environment.md`.

Key points:
* Go 1.21+ required
* OpenAPI-first with ogen code generation
* golangci-lint for code quality
* Standard Go toolchain (gofmt, go vet, go test)

---

## Quality Standards

All code must meet these standards before merging:

* **Tests Pass** - All existing and new tests must pass
* **Linting Clean** - No golangci-lint warnings
* **Formatted** - Code properly formatted with gofmt
* **Coverage** - Appropriate test coverage for new code
* **Documentation** - Public APIs documented with godoc comments
* **Integration** - Works correctly with existing codebase

---

*For core Go principles and development patterns, see `foundations.md`. For specific guidelines and checklists, see the `docs/guidelines/` directory.*
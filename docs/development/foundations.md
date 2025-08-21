# Go Project Development Foundations

This document outlines core principles, patterns, and workflows for Go development following Standard Go Project Layout.

---

## Project Structure

This project follows the Standard Go Project Layout:

```
project/
├── cmd/                    # Main applications  
├── internal/               # Private application and library code
├── pkg/                    # Library code safe for external use
├── api/                    # API definitions (OpenAPI, protobuf)
├── web/                    # Web application specific components
├── configs/                # Configuration file templates
├── docs/                   # Design and user documents
├── scripts/                # Scripts for build, install, analysis
├── test/                   # Additional external test apps and data
└── vendor/                 # Application dependencies (optional)
```

---

## Go Standards

### Language & Version
- **Go 1.21+** required
- **One module per repository**
- **Standard library preference** over external dependencies

### Interface Strategy
- **Define interfaces** in `internal/<domain>/`
- **Implement interfaces** in `internal/<domain>/v1/`

### Package & Naming
- **Short, lowercase names** - no underscores or mixedCaps
- **Clear, descriptive names** for functions and variables
- **Godoc comments** on all exported items

### Error Handling
- **Explicit error returns** - no exceptions
- **Wrap errors** with context using `fmt.Errorf("context: %w", err)`
- **Never ignore errors** with `_`

---

## Development Workflow

### Test-Driven Development (TDD)
1. **Write failing tests first**
2. **Write minimal code** to pass tests  
3. **Refactor** while keeping tests green
4. **Maintain 80%+ coverage** for new code

### Quality Gates
- **gofmt** - Proper formatting
- **golangci-lint** - Linting and conventions
- **go vet** - Static analysis
- **go test** - All tests passing

### Git Workflow
- **Atomic commits** - One logical change per commit
- **Conventional commit messages** - Clear, descriptive
- **Feature branches** for major changes
- **Clean history** before merging

---

## Code Patterns

### Context Usage
```go
// Always accept context as first parameter
func ServiceMethod(ctx context.Context, params Params) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // implementation
    }
}
```

### Interface Design
```go
// Define in internal/<domain>/
type Repository interface {
    Create(ctx context.Context, item Item) error
    GetByID(ctx context.Context, id string) (Item, error)
}

// Implement in internal/<domain>/v1/
type repository struct {
    db Database
}
```

### Error Handling
```go
func Process(data []byte) (*Result, error) {
    if len(data) == 0 {
        return nil, fmt.Errorf("empty data: %w", ErrInvalidInput)
    }
    // implementation
    return result, nil
}
```

### Table-Driven Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   Input
        want    Output
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Function() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## Anti-Patterns to Avoid

### Go-Specific
- **Goroutine leaks** - Always ensure exit paths
- **Mutex in structs** - Prefer channels for communication
- **Init() abuse** - Keep minimal and deterministic

### General
- **Scope creep** - Stay within declared boundaries
- **Assumption making** - Ask for clarification
- **Over-engineering** - Start simple, add complexity when needed
- **Test shortcuts** - Test all code, even "simple" parts

---

## Quality Standards

### Code Quality
- **Idiomatic Go** - Follow community conventions
- **Performance aware** - No unnecessary allocations
- **Concurrency safe** - Proper shared state handling
- **Comprehensive errors** - Appropriate error management

### Documentation Quality  
- **Clear purpose** - Write for humans
- **Complete coverage** - Explain the why
- **Runnable examples** - Include where helpful
- **Stay current** - Keep docs synced with code

### Change Quality
- **Backwards compatible** - Unless major version bump
- **Gradual deprecation** - Well-communicated deprecation path
- **Integration tested** - Works with existing codebase

---

## Success Metrics

### Universal
- **Quality gates pass** - All checks green
- **Test coverage** - Appropriate tests included
- **Documentation current** - Changes properly documented
- **Clean integration** - Works smoothly with existing code

### Go-Specific
- **Idiomatic code** - Follows Go conventions
- **No performance regressions** - Efficient implementations
- **Concurrency safety** - Thread-safe where needed
- **Proper error handling** - Comprehensive error management
- **Clean package design** - Focused responsibilities

---

## Reference Links

- **Go Style Guide**: https://google.github.io/styleguide/go/
- **Effective Go**: https://go.dev/doc/effective_go
- **Go Project Layout**: https://github.com/golang-standards/project-layout
- **Testing in Go**: https://go.dev/doc/tutorial/add-a-test
- **Go Modules**: https://go.dev/doc/modules/

---

*This document provides foundational knowledge for all developers working on Go projects. For specific implementation guidelines, see the related documentation in `docs/guidelines/` and `docs/architecture/`.*
# Go Coding Standards

This document defines Go-specific coding standards, patterns, and conventions for this project.

---

## Language Standards

### Version Requirements
- **Go 1.22+** required for all development
- **One module per repository**
- **Standard library preference** over external dependencies when functionality is equivalent

### Package Organization
- **Short, lowercase names** - no underscores or mixedCaps
- **Clear, descriptive names** for functions and variables
- **Godoc comments** on all exported items
- **Package comments** describing the package purpose

### Interface Design
Follow the standard interface-implementation pattern:
- **Define interfaces** in `internal/<domain>/interfaces.go`
- **Single implementations** directly in `internal/<domain>/`
- **Multiple implementations** in named subdirectories (e.g., `internal/<domain>/postgres/`)
- **Small interfaces** - prefer many small interfaces over large ones

---

## Code Patterns

### Context Usage
Always accept context as the first parameter for functions that:
- Make network calls
- Access databases
- Perform I/O operations
- May need cancellation or timeouts

```go
func ServiceMethod(ctx context.Context, params Params) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // implementation
    }
}
```

### Error Handling
- **Explicit error returns** - no exceptions or panics for business logic
- **Wrap errors** with context using `fmt.Errorf("context: %w", err)`
- **Never ignore errors** with `_`
- **Define sentinel errors** for expected error conditions

```go
var (
    ErrInvalidInput = errors.New("invalid input")
    ErrNotFound     = errors.New("not found")
)

func Process(data []byte) (*Result, error) {
    if len(data) == 0 {
        return nil, fmt.Errorf("empty data: %w", ErrInvalidInput)
    }
    
    result, err := processData(data)
    if err != nil {
        return nil, fmt.Errorf("processing data: %w", err)
    }
    
    return result, nil
}
```

### Interface Implementation
```go
// Define in internal/<domain>/interfaces.go
type Repository interface {
    Create(ctx context.Context, item Item) error
    GetByID(ctx context.Context, id string) (Item, error)
    Update(ctx context.Context, item Item) error
    Delete(ctx context.Context, id string) error
}

// Single implementation in internal/<domain>/repository.go
type repository struct {
    db Database
}

func NewRepository(db Database) Repository {
    return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, item Item) error {
    // implementation
}
```

### Constructor Pattern
- Use `New<TypeName>` for constructors
- Return interfaces, not concrete types when possible
- Accept configuration as options or structs

```go
type Config struct {
    Timeout time.Duration
    Retries int
}

func NewService(config Config) Service {
    return &service{
        timeout: config.Timeout,
        retries: config.Retries,
    }
}
```

---

## Concurrency Patterns

### Goroutine Management
- **Always ensure exit paths** for goroutines
- **Use context for cancellation**
- **Prefer channels over shared memory**

```go
func Worker(ctx context.Context, input <-chan Task, output chan<- Result) {
    defer close(output)
    
    for {
        select {
        case <-ctx.Done():
            return
        case task, ok := <-input:
            if !ok {
                return
            }
            result := process(task)
            select {
            case output <- result:
            case <-ctx.Done():
                return
            }
        }
    }
}
```

### Synchronization
- **Prefer channels** for communication between goroutines
- **Use mutexes** only when protecting shared state
- **Avoid global variables** - pass state explicitly

---

## Testing Standards

### Test Organization
- **Table-driven tests** for multiple test cases
- **Subtests** for logical grouping
- **Test helpers** for common setup/teardown

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   Input
        want    Output
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   Input{Value: "test"},
            want:    Output{Result: "processed"},
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   Input{Value: ""},
            want:    Output{},
            wantErr: true,
        },
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

### Test Naming
- **TestFunctionName** for unit tests
- **TestFunctionName_Condition** for specific scenarios
- **Use descriptive subtest names** in table-driven tests

### Mock and Stub Usage
- **Test against interfaces** when possible
- **Use dependency injection** for testability
- **Minimal mocking** - prefer real implementations when fast and reliable

---

## Documentation Standards

### Code Comments
- **Godoc comments** for all exported functions, types, and constants
- **Explain the why**, not just the what
- **Use complete sentences** starting with the item name

```go
// UserRepository provides methods for managing user data in the database.
// It handles user creation, retrieval, updates, and deletion operations.
type UserRepository interface {
    // Create adds a new user to the database.
    // Returns an error if the user already exists or if database operation fails.
    Create(ctx context.Context, user User) error
    
    // GetByID retrieves a user by their unique identifier.
    // Returns ErrUserNotFound if no user exists with the given ID.
    GetByID(ctx context.Context, id string) (User, error)
}
```

### Package Documentation
```go
// Package user provides functionality for managing user accounts,
// including authentication, profile management, and user preferences.
//
// The package follows the interface-implementation pattern with
// interfaces defined in this package and implementations in the same
// package for single implementations, or named subdirectories for
// multiple implementations.
package user
```

---

## Performance Guidelines

### Memory Management
- **Avoid unnecessary allocations** in hot paths
- **Reuse slices and maps** when possible
- **Use object pools** for frequently allocated objects
- **Profile before optimizing**

### Efficient Patterns
```go
// Prefer slice pre-allocation when size is known
items := make([]Item, 0, expectedSize)

// Use strings.Builder for string concatenation
var builder strings.Builder
builder.Grow(estimatedSize) // pre-allocate if size is known
for _, part := range parts {
    builder.WriteString(part)
}
result := builder.String()
```

---

## Anti-Patterns to Avoid

### Common Mistakes
- **Premature interfaces** - Wait for clear abstraction need
- **Goroutine leaks** - Always provide exit conditions
- **Ignoring errors** - Handle all error returns appropriately
- **Global state** - Pass dependencies explicitly
- **Init() abuse** - Keep initialization minimal and deterministic

### Code Smells
- **Large functions** - Break into smaller, focused functions
- **Deep nesting** - Use early returns and guard clauses
- **String concatenation in loops** - Use strings.Builder
- **Unmarked goroutines** - Always have a clear lifecycle

---

## Quality Checklist

Before submitting code, ensure:
- [ ] All exported items have godoc comments
- [ ] Error handling follows project conventions
- [ ] Tests cover happy path and error cases
- [ ] Code follows Go naming conventions
- [ ] Interfaces are minimal and focused
- [ ] Context is properly threaded through operations
- [ ] No goroutine leaks in concurrent code
- [ ] Performance considerations addressed for hot paths

---

## Reference

- **Effective Go**: https://go.dev/doc/effective_go
- **Go Code Review Comments**: https://github.com/golang/go/wiki/CodeReviewComments
- **Go Style Guide**: https://google.github.io/styleguide/go/

---

*For project-specific patterns and architecture, see `docs/architecture/`. For development workflows, see `docs/development/`.*
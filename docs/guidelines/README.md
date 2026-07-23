# Development Guidelines

This directory contains specific development guidelines and standards that ensure code quality, consistency, and maintainability across the Go Starter project.

## Purpose

Establish clear, actionable guidelines that development teams follow to maintain high code quality and consistent patterns throughout the codebase.

## Guidelines Overview

### [go-standards.md](go-standards.md)
**Go-specific coding standards and conventions**
- Naming conventions (packages, variables, functions)
- Code organization and structure
- Error handling patterns
- Documentation requirements
- Performance considerations
- Security best practices

### [interface-pattern.md](interface-pattern.md)
**Contract-first package and consumer-interface guidance**
- Package structure and organization
- Interface definition guidelines
- Versioned implementations
- Dependency injection patterns
- Testing strategies
- Migration approaches

### [testing-strategy.md](testing-strategy.md)
**Risk-based testing approach**
- Contract and regression test selection
- Unit and package testing patterns
- Integration and end-to-end boundary testing
- Coverage as a diagnostic signal
- Mocking and test doubles

## Quick Reference

### Naming Conventions
```go
// Packages: short, lowercase, no underscores
package auth

// Interfaces: noun + "er" suffix when possible
type Authenticator interface {}

// Exported: PascalCase
type UserService struct {}

// Unexported: camelCase
func validateToken() {}

// Constants: PascalCase or SCREAMING_SNAKE_CASE
const MaxRetries = 3
const DEFAULT_TIMEOUT = 30
```

### Error Handling
```go
// Always handle errors explicitly
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Define package-level errors
var (
    ErrNotFound = errors.New("resource not found")
    ErrInvalidInput = errors.New("invalid input")
)
```

### Testing Patterns
```go
// Table-driven tests
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Compliance Checklist

### Before Committing Code
- [ ] Code follows naming conventions
- [ ] All errors handled appropriately
- [ ] Public APIs have godoc comments
- [ ] Meaningful contract and risk tests pass
- [ ] No linting warnings (`make lint`)
- [ ] Code formatted (`make fmt`)

### Code Review Focus Areas
1. **Correctness** - Does it work as intended?
2. **Clarity** - Is the code easy to understand?
3. **Consistency** - Does it follow project patterns?
4. **Scope** - Is every behavior traceable to an approved requirement?
5. **Testing** - Does every test protect a meaningful contract or risk?
6. **Performance** - Are mechanisms and optimizations supported by evidence?
7. **Security** - Are boundaries and sensitive operations safe?

## Common Patterns

### Interface Definition
```go
// internal/auth/interfaces.go
package auth

type Service interface {
    Authenticate(ctx context.Context, token string) (*User, error)
    Authorize(ctx context.Context, user *User, resource string) error
}
```

### Implementation Structure
```go
// internal/auth/service.go
package auth

type service struct {
    repo auth.Repository
    logger Logger
}

func NewService(repo auth.Repository, logger Logger) auth.Service {
    return &service{
        repo: repo,
        logger: logger,
    }
}
```

### Constructor Pattern
```go
// Always use functional options for complex constructors
type Option func(*Config)

func WithTimeout(timeout time.Duration) Option {
    return func(c *Config) {
        c.Timeout = timeout
    }
}

func NewClient(opts ...Option) *Client {
    cfg := defaultConfig()
    for _, opt := range opts {
        opt(cfg)
    }
    return &Client{config: cfg}
}
```

## Anti-Patterns to Avoid

### ❌ Don't Do This
```go
// Naked returns
func calculate() (result int, err error) {
    result = 42
    return  // Unclear what's being returned
}

// Ignoring errors
result, _ := doSomething()  // Never ignore errors

// Global state
var globalConfig *Config  // Avoid global mutable state

// Interface pollution
type Doer interface {
    Do()  // Too generic
}
```

### ✅ Do This Instead
```go
// Explicit returns
func calculate() (int, error) {
    return 42, nil
}

// Handle errors
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed: %w", err)
}

// Dependency injection
func NewService(config *Config) *Service {
    return &Service{config: config}
}

// Specific interfaces
type UserAuthenticator interface {
    AuthenticateUser(ctx context.Context, credentials Credentials) (*User, error)
}
```

## Performance Guidelines

### Optimization Priorities
1. **Correctness first** - Make it work
2. **Clarity second** - Make it clear
3. **Performance third** - Make it fast (if needed)

### Common Optimizations
- Use `strings.Builder` for string concatenation
- Preallocate slices when size is known
- Use buffered channels appropriately
- Consider sync.Pool for frequently allocated objects
- Profile before optimizing

## Security Guidelines

### Always
- Validate all input
- Use parameterized queries
- Hash passwords with bcrypt
- Use crypto/rand for randomness
- Implement rate limiting
- Log security events

### Never
- Store secrets in code
- Trust user input
- Use MD5 or SHA1 for passwords
- Implement custom crypto
- Log sensitive data
- Ignore security warnings

## Documentation Standards

### Package Documentation
```go
// Package auth provides authentication and authorization functionality.
// It implements JWT-based authentication with refresh token support.
package auth
```

### Function Documentation
```go
// AuthenticateUser verifies the provided credentials and returns a user
// object if authentication succeeds. It returns ErrInvalidCredentials
// if the credentials are incorrect or ErrUserNotFound if the user
// doesn't exist.
func AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
    // implementation
}
```

## Related Documentation

- [../development/foundations.md](../development/foundations.md) - Core Go principles
- [../development/project-context.md](../development/project-context.md) - Project workflows
- [../architecture/](../architecture/) - Architecture decisions

## Enforcement

Guidelines are enforced through:
1. **Automated linting** - golangci-lint configuration
2. **Code review** - Team review checklist
3. **CI/CD pipeline** - Automated checks
4. **Documentation** - Clear examples and anti-patterns

---

*These guidelines ensure consistent, high-quality code that's maintainable and follows Go best practices.*

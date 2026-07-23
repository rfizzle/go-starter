# Interface Package Pattern

> **Scope:** Go package architecture and interface design patterns  
> **Applies to:** All Go packages in `internal/` and `pkg/` directories  
> **Version:** 1.0.0

## Overview

Define interfaces where behavior is consumed and keep them as small as the
consumer permits. Constructors accept those interfaces when a seam is needed
and return concrete types. Do not mirror a concrete provider with a broad
interface solely to make a hypothetical implementation swap possible.

## Purpose
- Enforce consistent, predictable package structure across Go project
- Enable clean separation between contract definition and implementation details
- Support real implementation seams and dependency inversion
- Provide testability through interface-based design
- Enable gradual migration between implementation versions

## Pattern Structure

```
internal/
└── <domain>/
    ├── interfaces.go         # Core domain interfaces (if multiple impls)
    ├── service.go           # Implementation (for single impl)
    ├── types.go             # Domain-specific types  
    ├── errors.go            # Domain-specific errors
    └── <provider>/          # Multiple implementations only
        └── client.go        # Provider-specific implementation
```

## Requirements

### Core Principles
- **Interface Segregation**: Interfaces are owned by the package that consumes the behavior
- **Implementation Isolation**: Single implementations in package root; multiple implementations in named subdirectories
- **Dependency Inversion**: Higher-level packages depend on interfaces, not concrete implementations
- **Single Responsibility**: Each interface defines cohesive set of related operations

### Specific Rules
1. Define an interface only when a consumer needs a behavioral seam
2. Supporting types and constants in `internal/<domain>/types.go`
3. Domain-specific errors in `internal/<domain>/errors.go`
4. For single implementation: directly in `internal/<domain>/`
5. Constructors accept narrow dependencies and return concrete types
6. Keep interfaces focused, normally one to three methods

## Implementation Guidelines

### Interface Definition

Define your domain interfaces in the main package directory:

```go
// internal/auth/interfaces.go
package auth

import "context"

// Service defines the core authentication operations
type Service interface {
    Authenticate(ctx context.Context, token string) (*User, error)
    Authorize(ctx context.Context, user *User, resource string) error
    RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
}

// Repository defines data persistence operations
type Repository interface {
    GetUser(ctx context.Context, id string) (*User, error)
    CreateUser(ctx context.Context, user *User) error
    UpdateLastLogin(ctx context.Context, userID string) error
}
```

### Single Implementation

For packages with one implementation, place it directly in the package:

```go
// internal/auth/service.go
package auth

import (
    "context"
    "github.com/rfizzle/go-starter/internal/auth"
)

type service struct {
    repo auth.Repository
    // ... other dependencies
}

// NewService creates a new auth service implementation
func NewService(repo auth.Repository) auth.Service {
    return &service{
        repo: repo,
    }
}

func (s *service) Authenticate(ctx context.Context, token string) (*auth.User, error) {
    // Implementation here
}
```

### Domain Types and Errors

Define shared types and errors at the package level:

```go
// internal/auth/types.go
package auth

import "time"

type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at"`
}
```

```go
// internal/auth/errors.go
package auth

import "errors"

var (
    ErrInvalidToken     = errors.New("invalid token")
    ErrTokenExpired     = errors.New("token expired")
    ErrUnauthorized     = errors.New("unauthorized")
    ErrUserNotFound     = errors.New("user not found")
)
```

### Dependency Injection

Use constructor functions that accept interfaces:

```go
// internal/auth/service.go
func NewService(
    repo auth.Repository,
    tokenManager TokenManager,
    logger Logger,
) auth.Service {
    return &service{
        repo:         repo,
        tokenManager: tokenManager,
        logger:       logger,
    }
}
```

### For Development Teams
- **Interface First**: Always define or reference interface before implementing
- **Implementation Strategy**: Use named subdirectories only when multiple implementations exist
- **Testing Strategy**: Write tests against interfaces, not concrete implementations
- **Documentation**: Document interface contracts thoroughly, including error conditions

### For Code Reviews
- **Contract Compliance**: Verify interface contracts followed and implementations properly isolated
- **Backward Compatibility**: Maintain backward compatibility or bump version appropriately
- **Integration Testing**: Ensure different implementations can be swapped without breaking consumers

### For Go Projects
- **Interface Composition**: Use interface embedding to compose larger interfaces from smaller ones
- **Type Assertions**: Minimize type assertions by designing interfaces appropriately
- **Error Handling**: Follow Go error handling conventions in all interface methods
- **Context Propagation**: Always thread context through interface methods for cancellation and timeouts

## Testing Strategy

### Interface Tests (Conformance)

Create conformance tests that validate any implementation:

```go
// internal/auth/conformance_test.go
package auth_test

import (
    "testing"
    "github.com/rfizzle/go-starter/internal/auth"
)

// TestServiceConformance tests any auth.Service implementation
func TestServiceConformance(t *testing.T, factory func() auth.Service) {
    service := factory()
    
    t.Run("Authenticate", func(t *testing.T) {
        // Test authentication logic
    })
    
    t.Run("Authorize", func(t *testing.T) {
        // Test authorization logic  
    })
}
```

### Implementation-Specific Tests

Test concrete implementations separately:

```go
// internal/auth/service_test.go
package auth_test

import (
    "testing"
    "github.com/rfizzle/go-starter/internal/auth"
)

func TestService(t *testing.T) {
    // Run conformance tests
    auth_test.TestServiceConformance(t, func() auth.Service {
        return auth.NewService(mockRepo, mockTokenManager, mockLogger)
    })
    
    // Add implementation-specific tests
    t.Run("SpecificBehavior", func(t *testing.T) {
        // Test specific functionality
    })
}
```

## Provider boundaries

Use provider subdirectories only when multiple implementations actually
exist. A provider constructor returns its concrete client. The consuming
package declares the behavior it needs:

```go
// internal/cli/seed/seed.go
type UserStore interface {
    ByEmail(ctx context.Context, email string) (domain.User, error)
    Create(ctx context.Context, params domain.CreateUserParams) (domain.User, error)
}

func New(users UserStore) (*Seeder, error) { ... }

// internal/app/app.go
users := user.NewRepository(pool)
seeder, err := seed.New(users)
```

The PostgreSQL package does not implement or return a broad database interface.
The concrete `*user.Repository` satisfies `seed.UserStore` structurally.

## Integration with Existing Architecture

### Controller Integration

Controllers should depend on interfaces, not implementations:

```go
// internal/controller/auth.go
type AuthController struct {
    authService auth.Service // Interface, not concrete type
}

func NewAuthController(authService auth.Service) *AuthController {
    return &AuthController{
        authService: authService,
    }
}
```

### Application Wiring

Wire up dependencies in the application layer:

```go
// internal/app/app.go
func (a *App) setupServices() {
    // Create repository implementations
    authRepo := postgres.NewAuthRepository(a.db)
    
    // Create service implementations  
    authService := auth.NewService(authRepo)
    
    // Create controllers
    authController := controller.NewAuthController(authService)
    
    // Register with web server
    a.webServer.RegisterAuthRoutes(authController)
}
```

## Compliance

### Checklist
- [ ] Domain interfaces defined in `internal/<domain>/interfaces.go`
- [ ] Types defined in `internal/<domain>/types.go`
- [ ] Errors defined in `internal/<domain>/errors.go`
- [ ] Implementation in appropriate location (root for single, subdirs for multiple)
- [ ] Constructor follows `New<InterfaceName>` pattern
- [ ] All methods accept `context.Context` as first parameter where appropriate
- [ ] Interface methods return explicit errors
- [ ] Implementation struct is unexported
- [ ] Tests written against interfaces, not concrete types
- [ ] Documentation covers interface contracts and error conditions

### Violations
- **Detection**: CI checks for package structure violations and interface compliance
- **Correction**: Refactor non-compliant code to follow interface-implementation pattern

## Exceptions
- **Simple Utility Packages**: Pure utility functions without state may not require interfaces
- **Generated Code**: API client code generated from OpenAPI specs may follow different patterns
- **Third-party Integrations**: Wrapper packages for external libraries may use adapted patterns

## Benefits of This Pattern

1. **Clean Abstraction**: Interfaces separate contract from implementation
2. **Testability**: Easy to mock interfaces for testing
3. **Versioning**: Multiple implementations can coexist
4. **Flexibility**: Easy to swap implementations
5. **Maintainability**: Clear separation of concerns
6. **Clean Architecture**: Clear separation between single and multiple implementations

## Examples in Codebase

Current examples of this pattern:
- `internal/cli/seed` - `UserStore` describes only the repository behavior seeding consumes
- `internal/controller` - `UserReader` describes only the lookup used by the HTTP handler
- `internal/cli` - lifecycle interfaces isolate command orchestration from the application

Apply this pattern when a consumer needs a test or implementation seam. Do not
add an interface to every exported concrete type.

## Success Metrics
- Interfaces remain small and consumer-owned.
- Constructors return concrete types.
- Unit tests substitute dependencies without mirroring provider internals.

## Related Documentation
- [docs/development/foundations.md](../development/foundations.md) - Go interface design patterns and core principles
- [docs/development/project-context.md](../development/project-context.md) - Project architecture overview
- [docs/guidelines/go-standards.md](../guidelines/go-standards.md) - Go-specific coding standards

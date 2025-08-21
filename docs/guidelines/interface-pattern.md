# Interface Package Pattern

> **Scope:** Go package architecture and interface design patterns  
> **Applies to:** All Go packages in `internal/` and `pkg/` directories  
> **Version:** 1.0.0

## Overview

We follow a pattern where **high-level interfaces** are defined at the package level (e.g., `internal/auth`) and **concrete implementations** are placed in versioned subfolders (e.g., `internal/auth/v1`). This approach provides clean abstraction boundaries, enables versioning of implementations, and maintains backward compatibility.

## Purpose
- Enforce consistent, predictable package structure across Go project
- Enable clean separation between contract definition and implementation details
- Support multiple implementations, versioning, and dependency inversion principles
- Provide testability through interface-based design
- Enable gradual migration between implementation versions

## Pattern Structure

```
internal/
└── <domain>/
    ├── interfaces.go          # Core domain interfaces and types
    ├── types.go              # Domain-specific types
    ├── errors.go             # Domain-specific errors
    ├── v1/                   # Version 1 implementations
    │   ├── service.go        # Core business logic implementation
    │   ├── handlers.go       # HTTP handlers (if applicable)
    │   └── models.go         # Version-specific models/DTOs
    ├── v2/                   # Version 2 implementations (future)
    └── <provider>/           # Provider-specific implementations
        └── implementation.go
```

## Requirements

### Core Principles
- **Interface Segregation**: Interfaces defined at top level of domain package, separate from implementation
- **Implementation Isolation**: All concrete implementations reside in versioned sub-packages (e.g., `v1/`, `v2/`)
- **Dependency Inversion**: Higher-level packages depend on interfaces, not concrete implementations
- **Single Responsibility**: Each interface defines cohesive set of related operations

### Specific Rules
1. Domain interfaces defined in `internal/<domain>/interfaces.go`
2. Supporting types and constants in `internal/<domain>/types.go`
3. Domain-specific errors in `internal/<domain>/errors.go`
4. Primary implementation in `internal/<domain>/v1/`
5. Constructor functions use pattern `New<InterfaceName>(dependencies...)`
6. Implementation structs should not be exported unless necessary

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

### Versioned Implementations

Place concrete implementations in versioned subdirectories:

```go
// internal/auth/v1/service.go
package v1

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
// internal/auth/v1/service.go
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
- **Version Awareness**: Place new implementations in appropriate version sub-packages
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
// internal/auth/v1/service_test.go
package v1_test

import (
    "testing"
    "github.com/rfizzle/go-starter/internal/auth"
    "github.com/rfizzle/go-starter/internal/auth/v1"
)

func TestService(t *testing.T) {
    // Run conformance tests
    auth_test.TestServiceConformance(t, func() auth.Service {
        return v1.NewService(mockRepo, mockTokenManager, mockLogger)
    })
    
    // Add v1-specific tests
    t.Run("V1_SpecificBehavior", func(t *testing.T) {
        // Test v1-specific functionality
    })
}
```

## Migration and Versioning

### Adding New Versions

When implementing breaking changes, create a new version:

```go
// internal/auth/v2/service.go - New implementation with breaking changes
package v2

// Enhanced interface with additional methods
func NewService(config Config) auth.Service {
    return &serviceV2{
        config: config,
    }
}
```

### Gradual Migration

Use feature flags or configuration to gradually migrate:

```go
// internal/app/app.go
func (a *App) setupAuth() auth.Service {
    if a.config.UseAuthV2 {
        return v2.NewService(a.config.AuthV2Config)
    }
    return v1.NewService(a.authRepo)
}
```

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
    authService := v1.NewService(authRepo)
    
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
- [ ] Implementation in versioned subpackage (`internal/<domain>/v1/`)
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
6. **Backward Compatibility**: Old versions remain available during migration

## Examples in Codebase

Current examples of this pattern:
- `internal/storage/` - Database abstraction with provider implementations
- `internal/auth/` - Authentication services (following this pattern)

This pattern should be used for all new domain packages and when refactoring existing code.

## Success Metrics
- **Interface Coverage**: >90% of public functionality exposed through interfaces—Target: 100%
- **Package Compliance**: Package structure follows pattern—Target: 100%
- **Test Coverage**: Interface-based tests vs implementation tests ratio—Target: 3:1

## Related Documentation
- [docs/development/foundations.md](../development/foundations.md) - Go interface design patterns and core principles
- [docs/development/project-context.md](../development/project-context.md) - Project architecture overview
- [docs/guidelines/go-standards.md](../guidelines/go-standards.md) - Go-specific coding standards
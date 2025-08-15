# Go Package Pattern Guide

This document describes the standardized package structure pattern used throughout this project for organizing interfaces and their implementations.

## Overview

We follow a pattern where **high-level interfaces** are defined at the package level (e.g., `internal/auth`) and **concrete implementations** are placed in versioned subfolders (e.g., `internal/auth/v1`). This approach provides clean abstraction boundaries, enables versioning of implementations, and maintains backward compatibility.

## Pattern Structure

```
internal/
└── <domain>/
    ├── interfaces.go          # Core domain interfaces and types
    ├── errors.go             # Domain-specific errors
    ├── v1/                   # Version 1 implementations
    │   ├── service.go        # Core business logic implementation
    │   ├── handlers.go       # HTTP handlers (if applicable)
    │   └── models.go         # Version-specific models/DTOs
    ├── v2/                   # Version 2 implementations (future)
    └── <provider>/           # Provider-specific implementations
        └── implementation.go
```

## Core Principles

### 1. Interface Definition at Package Level

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

### 2. Versioned Implementations

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

### 3. Domain Types and Errors

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

## Implementation Guidelines

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

### Provider-Specific Implementations

For external service integrations, create provider-specific subdirectories:

```go
// internal/auth/okta/client.go
package okta

import "github.com/rfizzle/go-starter/internal/auth"

type client struct {
    baseURL string
    apiKey  string
}

func NewClient(baseURL, apiKey string) auth.Service {
    return &client{
        baseURL: baseURL,
        apiKey:  apiKey,
    }
}
```

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
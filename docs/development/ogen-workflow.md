# Ogen Workflow Guide

Complete guide for OpenAPI-first development using ogen code generation in the Go Starter project.

## Overview

Ogen is a powerful OpenAPI 3.0 code generator for Go that creates type-safe, efficient HTTP clients and servers. This project uses ogen to generate API interfaces from the OpenAPI specification, ensuring the API contract drives the implementation.

## Key Concepts

### OpenAPI-First Development
1. **API specification is the source of truth** - All API changes start in `api/openapi.yml`
2. **Code generation creates interfaces** - Ogen generates Go interfaces from the spec
3. **Controllers implement interfaces** - Business logic implements generated interfaces
4. **Type safety throughout** - Strong typing from API to database

### Generated Code Structure

```
internal/api/v1/
├── oas_cfg_gen.go          # Configuration structures
├── oas_client_gen.go       # HTTP client implementation
├── oas_handlers_gen.go     # Handler interfaces
├── oas_interfaces_gen.go   # Core interfaces
├── oas_json_gen.go         # JSON marshaling/unmarshaling
├── oas_middleware_gen.go   # Middleware interfaces
├── oas_parameters_gen.go   # Request parameter types
├── oas_request_*_gen.go    # Request decoders
├── oas_response_*_gen.go   # Response encoders
├── oas_router_gen.go       # HTTP router
├── oas_schemas_gen.go      # Data models/DTOs
├── oas_server_gen.go       # Server implementation
├── oas_validators_gen.go   # Input validation
└── oas_*_gen.go           # Other generated files
```

## Workflow Steps

### 1. Modify OpenAPI Specification

Edit `api/openapi.yml` to define or update API endpoints:

```yaml
paths:
  /api/v1/users/{id}:
    get:
      operationId: getUser
      summary: Get user by ID
      tags:
        - users
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: User found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: User not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

components:
  schemas:
    User:
      type: object
      required:
        - id
        - email
        - name
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
          format: email
        name:
          type: string
        createdAt:
          type: string
          format: date-time
```

### 2. Generate Code

Run the generation command:

```bash
# Using Make
make generate

# Or directly with go generate
go generate ./...

# Or using ogen directly
go run github.com/ogen-go/ogen/cmd/ogen \
  --target internal/api/v1 \
  --package api \
  --clean \
  --convenient-errors \
  api/openapi.yml
```

### 3. Implement Controller

Create or update controller to implement the generated interface:

```go
// internal/controller/user.go
package controller

import (
    "context"
    "github.com/rfizzle/go-starter/internal/api/v1"
    "github.com/rfizzle/go-starter/internal/storage/driver"
)

type UserController struct {
    db driver.DB
}

func NewUserController(db driver.DB) *UserController {
    return &UserController{db: db}
}

// GetUser implements the generated interface method
func (c *UserController) GetUser(ctx context.Context, params api.GetUserParams) (api.GetUserRes, error) {
    user, err := c.db.GetUser(ctx, params.ID)
    if err != nil {
        if errors.Is(err, driver.ErrNotFound) {
            return &api.GetUserNotFound{
                Message: "User not found",
            }, nil
        }
        return nil, err
    }
    
    return &api.User{
        ID:        user.ID,
        Email:     user.Email,
        Name:      user.Name,
        CreatedAt: user.CreatedAt,
    }, nil
}
```

### 4. Wire Up in Application

Register the controller with the generated server:

```go
// internal/app/app.go
func (a *App) setupRoutes() {
    // Create controller
    userController := controller.NewUserController(a.db)
    
    // Create ogen server with our controller
    srv, err := api.NewServer(userController)
    if err != nil {
        log.Fatal(err)
    }
    
    // Mount to Chi router
    a.router.Mount("/", srv)
}
```

## Common Patterns

### Request Validation

Ogen automatically validates requests based on the OpenAPI schema:

```yaml
parameters:
  - name: limit
    in: query
    schema:
      type: integer
      minimum: 1
      maximum: 100
      default: 10
```

The generated code will:
- Validate the parameter is an integer
- Ensure it's between 1 and 100
- Use default value if not provided

### Error Responses

Define error schemas in OpenAPI:

```yaml
components:
  schemas:
    Error:
      type: object
      required:
        - message
        - code
      properties:
        message:
          type: string
        code:
          type: string
        details:
          type: object
```

Implement in controller:

```go
func (c *Controller) SomeOperation(ctx context.Context, req api.Request) (api.Response, error) {
    if err := c.validate(req); err != nil {
        return &api.ErrorBadRequest{
            Message: "Validation failed",
            Code:    "VALIDATION_ERROR",
            Details: api.OptObject{
                Set: true,
                Value: map[string]interface{}{
                    "field": "email",
                    "error": "invalid format",
                },
            },
        }, nil
    }
    // ... rest of implementation
}
```

### Optional Fields

Use ogen's optional types for nullable fields:

```go
type UpdateUserReq struct {
    Name  api.OptString `json:"name"`
    Email api.OptString `json:"email"`
}

// In controller
if req.Name.Set {
    user.Name = req.Name.Value
}
```

### Streaming Responses

For large responses, use streaming:

```yaml
responses:
  '200':
    content:
      application/x-ndjson:
        schema:
          type: string
          format: binary
```

## Testing Generated Code

### Unit Testing Controllers

```go
func TestGetUser(t *testing.T) {
    mockDB := &mockDriver.DB{}
    controller := NewUserController(mockDB)
    
    mockDB.On("GetUser", mock.Anything, "123").Return(&driver.User{
        ID:    "123",
        Email: "test@example.com",
        Name:  "Test User",
    }, nil)
    
    res, err := controller.GetUser(context.Background(), api.GetUserParams{
        ID: "123",
    })
    
    require.NoError(t, err)
    user, ok := res.(*api.User)
    require.True(t, ok)
    assert.Equal(t, "test@example.com", user.Email)
}
```

### Integration Testing

```go
func TestUserAPI(t *testing.T) {
    // Create test server
    srv := setupTestServer(t)
    
    // Create ogen client
    client, err := api.NewClient(srv.URL)
    require.NoError(t, err)
    
    // Test API call
    user, err := client.GetUser(context.Background(), api.GetUserParams{
        ID: "123",
    })
    require.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}
```

## Best Practices

### 1. Schema Design
- Use `$ref` for reusable schemas
- Define clear validation rules
- Use appropriate formats (uuid, email, date-time)
- Document all endpoints and schemas

### 2. Controller Implementation
- Keep controllers thin - delegate to services
- Handle all error cases explicitly
- Use context for cancellation
- Return appropriate HTTP status codes

### 3. Code Organization
```
internal/
├── api/v1/          # Generated code (don't edit)
├── controller/      # Controller implementations
│   ├── user.go
│   ├── auth.go
│   └── ...
├── service/         # Business logic
└── storage/         # Data layer
```

### 4. Versioning
- Version your API in the path (`/api/v1/`)
- Keep old versions during migration
- Use feature flags for gradual rollout

## Troubleshooting

### Common Issues

#### Generation Fails
```bash
# Check OpenAPI spec is valid
npx @openapitools/openapi-generator-cli validate -i api/openapi.yml

# Check ogen version
go list -m github.com/ogen-go/ogen
```

#### Type Mismatch
- Ensure OpenAPI types match Go types
- Use correct formats (int32, int64, float, double)
- Check optional vs required fields

#### Missing Implementation
```
controller does not implement api.Handler
```
- Run `make generate` after spec changes
- Implement all required methods
- Check method signatures match exactly

### Debugging Tips

1. **Check generated interfaces**
   ```bash
   grep "type Handler interface" internal/api/v1/oas_interfaces_gen.go
   ```

2. **Validate implementation**
   ```go
   var _ api.Handler = (*Controller)(nil) // Compile-time check
   ```

3. **Review generated schemas**
   - Check `oas_schemas_gen.go` for type definitions
   - Verify optional types are handled correctly

## Advanced Topics

### Custom Middleware

```go
func loggingMiddleware(next api.Handler) api.Handler {
    return api.HandlerFunc(func(ctx context.Context, req interface{}) (interface{}, error) {
        start := time.Now()
        res, err := next.Handle(ctx, req)
        log.Printf("Request took %v", time.Since(start))
        return res, err
    })
}
```

### Security Schemes

```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - bearerAuth: []
```

### Custom Validation

```go
func (c *Controller) ValidateRequest(req api.Request) error {
    // Custom business logic validation
    if req.Amount > c.maxAmount {
        return errors.New("amount exceeds limit")
    }
    return nil
}
```

## Related Documentation

- [Ogen Documentation](https://ogen.dev)
- [OpenAPI Specification](https://swagger.io/specification/)
- [project-context.md](project-context.md) - Project overview
- [interface-pattern.md](../guidelines/interface-pattern.md) - Package patterns

---

*This guide ensures consistent, type-safe API development using ogen code generation from OpenAPI specifications.*
# Development Environment

This document covers development environment setup, tooling requirements, and troubleshooting for the Go Starter project.

## Prerequisites

### Go Version
- **Go 1.22+** is required
- **Latest stable version** is recommended
- Use `go version` to check your current version

```bash
go version
# Should show: go version go1.22.x platform/arch
```

### Development Tools

#### Required Tools
```bash
# Install golangci-lint for code quality
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install ogen for OpenAPI code generation
go install github.com/ogen-go/ogen/cmd/ogen@latest

# Install goimports for import management
go install golang.org/x/tools/cmd/goimports@latest

# Install yamlfmt for YAML formatting (optional but recommended)
go install github.com/google/yamlfmt/cmd/yamlfmt@latest
```

#### Development Database
```bash
# PostgreSQL 15+ (via Docker recommended)
docker run --name go-starter-postgres \
  -e POSTGRES_PASSWORD=devpassword \
  -e POSTGRES_DB=gostarter \
  -p 5432:5432 -d postgres:15
```

## Project Setup

### Initial Setup
```bash
# Clone and setup project
git clone <repository-url>
cd go-starter

# Install dependencies
go mod tidy

# Generate API code
make generate

# Verify everything works
make ci
```

### Environment Configuration
```bash
# Copy example environment file
cp .env.example .env

# Edit with your local settings
vim .env
```

### Generated Code

The project uses ogen for OpenAPI code generation:

```bash
# Generate server/client/types into an internal package
go run github.com/ogen-go/ogen/cmd/ogen -convenient-errors --clean --package api --target internal/api/v1 api/openapi.yml
```

### Typical integration

* Generated interfaces/types: `internal/api/v1/...`
* Implement server methods in `internal/controller/...` where handlers satisfy the generated `Server` interface.
* Compose middlewares (auth/logging/metrics) in `internal/http`.

## Build System

### Makefile Commands

The project uses a Makefile for all common operations:

```bash
# Code generation and formatting
make generate    # Regenerate OpenAPI code
make fmt         # Format Go and YAML files
make tidy        # Clean up go.mod

# Quality checks
make lint        # Run golangci-lint
make vet         # Run go vet
make test        # Run tests
make race        # Run tests with race detector

# Building
make build       # Multi-platform builds
make linux       # Linux build only  
make osx         # macOS build only
make windows     # Windows build only

# Development workflow
make ci          # Full CI pipeline (generate, fmt, lint, test)
make clean       # Clean build artifacts
make docker      # Build Docker image
```

### NPM Script Alternative
```bash
# The project also supports npm-style commands
npm run build
npm run test  
npm run fmt
npm run lint
```

## IDE Setup

### VS Code
Recommended extensions:
- Go (official Go extension)
- YAML (Red Hat)
- OpenAPI (Swagger) Viewer

Settings (`.vscode/settings.json`):
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.useLanguageServer": true
}
```

### GoLand/IntelliJ
- Enable Go modules support
- Set goimports as formatter
- Configure golangci-lint as external tool

## Quality Standards

### Code Quality
All code must pass:
```bash
make fmt         # Code formatting
make lint        # Code linting  
make vet         # Static analysis
make test        # All tests
```

### Pre-commit Hooks (Recommended)
```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install

# Run on all files
pre-commit run --all-files
```

## Troubleshooting

### Common Issues

#### Build Failures
```bash
# Clean and rebuild
make clean
go mod tidy
make generate
make build
```

#### Test Failures
```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -v ./internal/auth -run TestService

# Run with race detector
make race
```

#### Generated Code Issues
```bash
# Clean generated files
rm -rf internal/api/v1/*_gen.go

# Regenerate
make generate
```

#### Database Issues
```bash
# Reset database
docker stop go-starter-postgres
docker rm go-starter-postgres
# Run initial setup command again
```

### Performance Issues

#### Slow Tests
```bash
# Run tests in parallel
go test -v -parallel 4 ./...

# Skip integration tests
go test -short ./...
```

#### Slow Builds
```bash
# Use build cache
export GOCACHE=$(go env GOCACHE)

# Parallel builds
make build -j4
```

### Environment Variables

#### Required Variables
```bash
DATABASE_URL=postgres://user:pass@localhost/gostarter
PORT=8080
LOG_LEVEL=info
```

#### Optional Variables
```bash
ENABLE_PPROF=true
CORS_ORIGINS=*
JWT_SECRET=your-secret-key
```

## Docker Development

### Development Container
```bash
# Build development image
docker build -f Dockerfile.dev -t go-starter:dev .

# Run with hot reload
docker run -v $(pwd):/app -p 8080:8080 go-starter:dev
```

### Production Container
```bash
# Build production image
make docker

# Run production container
docker run -p 8080:8080 go-starter:latest
```

## Testing

### Test Categories
- **Unit Tests**: Fast, isolated tests
- **Integration Tests**: Database and external service tests  
- **E2E Tests**: Full application tests

### Test Commands
```bash
# All tests
make test

# Unit tests only
go test -short ./...

# Integration tests
go test -tags=integration ./...

# Specific package
go test ./internal/auth

# With coverage
go test -cover ./...
```

### Test Database
```bash
# Use separate test database
export TEST_DATABASE_URL=postgres://user:pass@localhost/gostarter_test

# Or use in-memory database for faster tests
export USE_MEMORY_DB=true
```

## Performance Profiling

### CPU Profiling
```bash
# Start server with profiling
ENABLE_PPROF=true go run main.go

# Generate CPU profile
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory profile
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Benchmarks
```bash
# Run benchmarks
go test -bench=. ./...

# With memory allocation stats
go test -bench=. -benchmem ./...
```

---

*This development environment setup ensures consistent, high-quality development across all contributors to the Go Starter project.*
# Tech Stack

## Backend
- Language: Go (latest stable)
- HTTP Router: go-chi with CORS middleware
- API: OpenAPI 3.x specification
- Code Generation: ogen for server/client code

## Frontend
- Framework: React with functional components and hooks
- State Management: React hooks (useState, useContext, useReducer)
- Styling: Tailwind CSS

## Database
- Primary: PostgreSQL
- Driver: pgx
- Utilities: scany for struct scanning
- Caching: Redis

## CLI
- Framework: Cobra

## Configuration
- Library: Viper (supports YAML, JSON, env variables)

## Observability
- Logging: Zap (structured logging)
- Metrics: Prometheus client
- Tracing: OpenTelemetry

## Testing
- Framework: testify (assert, require, mock)
- Database Mocking: pgxmock

## Infrastructure
- Containerization: Docker
- Orchestration: Kubernetes
- CI/CD: GitLab CI

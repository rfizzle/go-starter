# Tech Stack

This document outlines the technology stack used in this project, based on its dependencies.

## Language

- **Go**: The primary programming language for this project.

## Web & API

- **Chi**: A lightweight, idiomatic and composable router for building Go HTTP services.
- **OpenAPI (Swagger)**: The project utilizes the OpenAPI specification for its API, with `ogen` for code generation and `go-openapi` for validation and other utilities.
- **CORS**: Middleware for handling Cross-Origin Resource Sharing in the Chi router.

## Database

- **PostgreSQL**: The project is configured to use PostgreSQL as its database.
- **pgx**: The Go driver and toolkit used for interacting with PostgreSQL.
- **scany**: A library to simplify scanning database rows into Go structs.

## CLI

- **Cobra**: A library for creating powerful and modern command-line interfaces.

## Configuration

- **Viper**: A complete configuration solution for Go applications, handling various formats and sources.

## Observability

- **Zap**: A fast, structured, leveled logging library.
- **OpenTelemetry**: Provides instrumentation for tracing and metrics to observe application performance.
- **Prometheus**: The Go client library is used for exposing application metrics in a Prometheus-compatible format.

## Testing

- **testify**: A toolkit providing assertion and mocking capabilities for tests.
- **pgxmock**: A mock driver for `pgx` to test database interactions without a running PostgreSQL instance.

## Validation

- **gookit/validate**: A library for data validation.
- **go-openapi/validate**: Used for validating data against the OpenAPI schema.

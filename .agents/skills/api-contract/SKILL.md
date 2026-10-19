---
name: api-contract
description: How the HTTP API is changed in this template — edit api/openapi.yml, regenerate the ogen transport in internal/api/v1 with make generate, then adapt the mapping in internal/controller (and the security handler with internal/auth). Use when adding or changing an operation, schema, status code, error shape or security scheme, when the controller no longer satisfies the generated interface, or when make generate-check reports drift under internal/api/v1. Not for package placement (go-project-layout), persistence (golang-database) or business logic in internal/modules.
license: MIT
metadata:
  author: rfizzle
  standard: agentskills.io
---

# API Contract

`api/openapi.yml` is the contract; the Go transport under `internal/api/v1/`
is derived from it by ogen and is never edited. A change starts in the YAML,
is regenerated, then adapted at the edge.

## Who owns what

| Path | Owns |
|---|---|
| `api/openapi.yml` | Paths, operations, schemas, security, status codes, public error shapes |
| `generate.go` | The pinned `go generate` invocation of ogen |
| `internal/api/v1/` | ogen output; never edited |
| `internal/controller/` | Mapping generated requests and responses to domain calls; consumer-side ports |
| `internal/auth/` | Token validation behind the generated security handler |
| `internal/app/` | Building the concretes and assembling the generated server |
| `internal/platform/http/server/` | Chi mounting, middleware, reserved routes, server lifecycle |

Business logic belongs in `internal/modules/<capability>/`, never in generated
transport or the controller.

## Change the contract

1. Edit `api/openapi.yml`.
2. `make generate` (runs ogen through `go generate`, and sqlc). ogen is a
   module dependency pinned in `go.mod`; never install or run a separate
   ogen binary.
3. Read the generated interface and schema delta: `git diff -- internal/api/v1`.
4. Implement the changed handler or security method in `internal/controller/`
   (token checks delegate to `internal/auth/`).
5. Update the focused controller or HTTP tests for the public contract.
6. `make generate-check`, then `make check`.

## Mapping rules

- Generated `api` types stay in the controller, auth, the HTTP server and the
  composition root (`make architecture-check` enforces it).
- Convert business values to response schemas in the controller.
- Map expected business errors to the bounded response the contract declares.
- Return unexpected errors so the server's error path logs them; never put a
  raw error in a response body.
- Let ogen enforce schema-level request validation.
- The controller declares small interfaces for only the behaviour it consumes;
  the concretes are injected from `internal/app`, and the controller never
  imports a concrete module.
- Add a compile-time assertion (`var _ api.X = (*T)(nil)`) where an edge must
  satisfy a generated interface.

Security schemes are declared in the YAML. A public operation (login) declares
no security; a protected one names the bearer scheme. Login rate limiting wraps
the generated handler in the composition root and answers with the 429 the
contract declares.

## Test

Do not test ogen's generated code. Test what the project owns:

- controller unit tests for mapping, bounded errors and the consumed ports;
- HTTP tests for status, headers, request validation and security;
- application tests for server assembly and the protected round trip;
- `make generate-check` for source and output agreeing.

Run `make test-integration` when the change depends on real PostgreSQL, and
`make race` when it touches lifecycle or middleware.

## When generation surprises you

- More changed than expected: confirm only `api/openapi.yml`, `generate.go`,
  `go.mod` or generator dependencies changed, `make generate` again, and read
  `git diff -- internal/api/v1`.
- The controller no longer compiles against the contract: find the generated
  `Handler` interface under `internal/api/v1/` and run
  `go test ./internal/controller`.
- `make generate-check` reports drift: regenerate from the source. Never patch
  an `oas_*_gen.go` file by hand.

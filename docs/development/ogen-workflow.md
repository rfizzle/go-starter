# Ogen Workflow Guide

Go Starter uses ogen to generate the HTTP transport from
`api/openapi.yml`. The OpenAPI document is the contract; generated Go files are
derived output.

## Ownership boundary

| Path | Responsibility |
| --- | --- |
| `api/openapi.yml` | Paths, operations, schemas, security, status codes, and public error shapes |
| `generate.go` | Pinned `go generate` invocation |
| `internal/api/v1/` | ogen output; never edit directly |
| `internal/controller/` | Generated request/response mapping and consumer-side ports |
| `internal/auth/` | JWT validation and generated security-handler integration |
| `internal/app/` | Concrete dependency construction and generated server assembly |
| `internal/platform/http/server/` | Chi mounting, middleware, reserved routes, and server lifecycle |

Business logic belongs in `internal/modules/<capability>/`, not in generated
transport or controllers.

## Change an API contract

1. Edit `api/openapi.yml`.
2. Run `make generate`.
3. Review the generated interface and schema delta under `internal/api/v1/`.
4. Implement the changed handler or security method at the transport edge.
5. Update focused controller/HTTP tests for the public contract.
6. Run:

```sh
make generate-check
make check
```

`make generate` runs both ogen and sqlc. The ogen invocation is:

```sh
go generate
```

The directive in `generate.go` runs the module-pinned ogen version:

```text
go run github.com/ogen-go/ogen/cmd/ogen --clean --package api \
  --target internal/api/v1 api/openapi.yml
```

Do not install or invoke a separate unpinned ogen binary.

## Current contract

The committed OpenAPI document defines four operations:

| Operation | Purpose |
| --- | --- |
| `HealthLiveness` | Process-oriented liveness |
| `HealthReadiness` | Dependency-aware readiness |
| `Login` | Credential verification and JWT issuance |
| `GetCurrentUser` | Bearer-protected current-user response |

`internal/controller.Controller` implements the generated handler and security
contracts. Small interfaces such as `UserReader`, `ReadinessChecker`, and
`Authenticator` describe only behavior the edge consumes. Concrete module and
platform implementations are injected from `internal/app`.

## Mapping rules

- Keep generated `api.*` types inside controller, auth, HTTP server, and
  composition-root edges.
- Convert business values to response schemas in `internal/controller/`.
- Map expected business errors to the bounded response declared in OpenAPI.
- Return unexpected internal errors so the server's error path can handle and
  log them; do not expose raw errors in response bodies.
- Let ogen enforce schema-level request validation.
- Keep credential verification, persistence, and other business rules in
  modules.
- Add compile-time assertions when an edge must satisfy a generated interface.

The login flow demonstrates the intended direction:

```text
OpenAPI Login
  → generated api.Handler
  → controller.Login
  → controller.Authenticator
  → app adapter
  → modules/credential.Service
```

The controller does not import the concrete credential module.

## Security

Security schemes are declared in `api/openapi.yml`. The generated security
handler is implemented in `internal/controller/security.go` and delegates JWT
validation to `internal/auth`.

The login operation is intentionally public. The current-user operation
requires the `BearerAuth` scheme. A token issued by login uses the configured
issuer, audience, TTL, and subject and is accepted by the same server's
protected operation.

Login rate limiting is installed around the generated API handler in the
composition root. It derives client identity through configured trusted proxy
CIDRs and returns the OpenAPI-declared 429 response.

## Testing

Do not test ogen's generated implementation. Test project-owned behavior:

- controller unit tests for mapping, bounded errors, and consumer contracts;
- HTTP tests for status, headers, request validation, and security behavior;
- application tests for generated-server assembly, login, `/me`, listener
  lifecycle, and shutdown; and
- `make generate-check` for source/output consistency.

Use `make test-all` when the API change depends on real PostgreSQL behavior and
`make race` when it changes lifecycle or middleware.

## Troubleshooting

If generation changes more than expected:

1. confirm only `api/openapi.yml`, `generate.go`, `go.mod`, or generator
   dependencies changed;
2. run `make generate` again;
3. inspect `git diff -- internal/api/v1`; and
4. run `make generate-check`.

If the controller no longer implements the generated contract, inspect:

```sh
rg 'type Handler interface' internal/api/v1
go test ./internal/controller
```

If `make generate-check` reports drift, regenerate from the source. Never patch
an `oas_*_gen.go` file by hand.

Related guidance:

- [OpenAPI specification](https://spec.openapis.org/oas/latest.html)
- [ogen documentation](https://ogen.dev/)
- [project context](project-context.md)
- [module/platform boundaries](../architecture/module-platform-boundaries.md)

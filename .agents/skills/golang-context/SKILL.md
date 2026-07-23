---
name: golang-context
description: "Idiomatic context.Context usage in Golang—creation, propagation, cancellation, timeouts, deadlines, values, and cross-service tracing. Use when designing or reviewing context-aware APIs, cancellation ownership, deadline behavior, request-scoped values, or context-related leaks. Not for code that merely forwards an existing context without changing its behavior."
license: MIT
---


# Go context.Context Best Practices

`context.Context` is Go's mechanism for propagating cancellation signals, deadlines, and request-scoped values across API boundaries and between goroutines. Think of it as the "session" of a request — it ties together every operation that belongs to the same unit of work.

## Best Practices Summary

1. The same context MUST be propagated through the entire request lifecycle: HTTP handler → service → DB → external APIs
2. `ctx` MUST be the first parameter, named `ctx context.Context`
3. NEVER store context in a struct — pass explicitly through function parameters
4. NEVER pass `nil` context — use `context.TODO()` if unsure
5. Every cancel function MUST be called on all paths. An immediate `defer cancel()` is
   usually right in a bounded function; loops and ownership transfer need explicit scope.
6. `context.Background()` MUST only be used at the top level (main, init, tests)
7. **Use `context.TODO()`** as a placeholder when you know a context is needed but don't have one yet
8. NEVER create a new `context.Background()` in the middle of a request path
9. Context value keys MUST be unexported types to prevent collisions
10. Context values MUST only carry request-scoped metadata — NEVER function parameters
11. Use `context.WithoutCancel` only when work intentionally outlives a request and has a
    new owner, explicit timeout/cancellation, and shutdown tracking. It removes the
    parent's deadline as well as cancellation.

## Creating Contexts

| Situation | Use |
| --- | --- |
| Entry point (main, init, test) | `context.Background()` |
| Function needs context but caller doesn't provide one yet | `context.TODO()` |
| Inside an HTTP handler | `r.Context()` |
| Need cancellation control | `context.WithCancel(parentCtx)` |
| Need a deadline/timeout | `context.WithTimeout(parentCtx, duration)` |

## Context Propagation: The Core Principle

The most important rule: **propagate the same context through the entire call chain**. When you propagate correctly, cancelling the parent context cancels all downstream work automatically.

```go
// ✗ Bad — creates a new context, breaking the chain
func (s *OrderService) Create(ctx context.Context, order Order) error {
    _, err := s.db.Exec(context.Background(), "INSERT INTO orders ...", order.ID)
    return err
}

// ✓ Good — propagates the caller's context
func (s *OrderService) Create(ctx context.Context, order Order) error {
    _, err := s.db.Exec(ctx, "INSERT INTO orders ...", order.ID)
    return err
}
```

## Deep Dives

- **[Cancellation, Timeouts & Deadlines](./references/cancellation.md)** — How cancellation propagates: `WithCancel` for manual cancellation, `WithTimeout` for automatic cancellation after a duration, `WithDeadline` for absolute time deadlines. Patterns for listening (`<-ctx.Done()`) in concurrent code, `AfterFunc` callbacks, and `WithoutCancel` for operations that must outlive their parent request (e.g., audit logs).

- **[Context Values & Cross-Service Tracing](./references/values-tracing.md)** — Safe context value patterns: unexported key types to prevent namespace collisions, when to use context values (request ID, user ID) vs function parameters. Trace context propagation: OpenTelemetry trace headers, correlation IDs for log aggregation, and marshaling/unmarshaling context across service boundaries.

- **[Context in HTTP Servers & Service Calls](./references/http-services.md)** — HTTP handler context: `r.Context()` for request-scoped cancellation, middleware integration, and propagating to services. HTTP client patterns: `NewRequestWithContext`, client timeouts, and retries with context awareness. Database operations: pass context to pgx `Query`, `QueryRow`, and `Exec`.

## Cross-References

- → See the `golang-concurrency` skill for goroutine cancellation patterns using context
- → See the `golang-database` skill for context-aware pgx database operations
- → See the `golang-observability` skill for trace context propagation with OpenTelemetry
- → See the `golang-design-patterns` skill for timeout and resilience patterns

## Enforce with Linters

Many context pitfalls are caught by `govet` and `staticcheck`; use the repository's
`make lint` target so the configured versions and settings are applied.

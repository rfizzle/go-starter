# Dependency health

This package evaluates the named technical dependencies required for the
service to accept traffic. A single timeout bounds the complete evaluation,
and checks must honor the context they receive.

The aggregator returns the first failing check's stable name wrapping the
underlying error for internal startup diagnostics; it keeps no state and logs
nothing, leaving readiness transition logging to the application lifecycle.
HTTP adapters consume the boolean readiness result and must not expose raw
dependency errors.

Liveness is intentionally outside this package: it remains process-oriented
and does not call PostgreSQL or any other external service.

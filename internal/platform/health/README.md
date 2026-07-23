# Dependency health

This package evaluates the named technical dependencies required for the
service to accept traffic. A single timeout bounds the complete evaluation,
and checks must honor the context they receive.

The aggregator retains underlying errors for internal startup diagnostics, but
logs only stable check names and readiness transitions. HTTP adapters consume
the boolean readiness result and must not expose raw dependency errors.

Liveness is intentionally outside this package: it remains process-oriented
and does not call PostgreSQL or any other external service.

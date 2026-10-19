# Architecture

The big shapes of the service and why they are those shapes. Where each
package lives and which way it may import is in `AGENTS.md`; the boundary
decision is in [decisions.md](decisions.md#d1-module-and-platform-boundaries).

Go Starter is an OpenAPI-first HTTP service with explicit composition,
PostgreSQL persistence, token authentication and a module/platform package
split.

## Composition

One composition root builds every concrete dependency and owns startup and
shutdown ordering; everything it builds is handed to consumers through small
interfaces the consumers declare. Commands that need less than the whole
service (seeding, migrating) are smaller composition roots of their own rather
than reaching into the full one. Business and domain packages never construct
loggers, exporters or connections. Shutdown drains in-flight requests before
the database pool closes, and cancellation reaches every running component.

## Transport

`api/openapi.yml` is the contract. The server and types are generated from it
and are never edited, so the controller is the only place that maps between
transport and domain, and it contains no business logic. Generated types stay
at the edge so a contract change cannot leak into the business packages.

Beside the contract the server exposes metrics, profiling (off by default and
outside the contract) and the embedded static site, which is served only
outside the reserved API, metrics and profiling paths.

Login is rate limited per client address by a bounded, in-memory limiter, so
each replica enforces its own limit rather than a global one. Forwarding
headers are believed only when the immediate peer is inside a configured
trusted-proxy range, walking from that peer toward the client and stopping at
the first untrusted address; with no range configured, the peer itself is the
key.

## Persistence

The service owns one connection pool, shared by startup checks, metrics,
migrations and the repositories. Migrations are serialized across instances
with a database-level lock, and applying them at startup is off by default in
favour of a one-shot migrate command, so a rolling deploy does not race
itself. Fixed queries are generated from SQL; dynamic ones are built in code.
Rows are mapped to domain data, which deliberately cannot carry a password
hash. How persistence code is written is the `golang-database` skill.

## Configuration and observability

Configuration is decoded and validated in a package with no side effects;
reading files and the environment belongs to each command's composition root.
Flags override `GOSTARTER_*` environment variables, which override an
explicitly selected configuration file, and validation finishes before any
listener or database work begins. Trace and metric export is off unless
enabled with a collector.
Logging and tracing providers are built once and passed down. Metrics and
profiling are HTTP and database platform concerns, not business ones.

## Health

Readiness runs named dependency checks inside one bounded time budget and
combines them with the application's lifecycle state, so a draining instance
reports not ready. Liveness is about the process only and never touches the
database, so a database outage does not restart healthy instances.

## Verification

- `make architecture-check` verifies package placement and import direction.
- `make generate-check` detects stale generated output.
- `make test` is the default suite with no external service;
  `make test-integration` covers the disposable external boundaries.
- `make race` is required for concurrency, lifecycle, middleware or database
  changes.

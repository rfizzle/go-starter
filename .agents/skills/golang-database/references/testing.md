# Testing PostgreSQL Repositories

## Contents

- [Unit tests](#unit-tests)
- [Service tests with fakes](#service-tests-with-fakes)
- [Integration tests](#integration-tests)

## Unit tests

Unit-test what needs no database: row-to-domain mapping (that persistence-only
fields are dropped), identifier parsing and query-builder inputs. They run in
`make test`.

## Service tests with fakes

Business services receive a local consumer port. Use a small hand-written fake
for it; do not mock pgx underneath business logic.

## Integration tests

SQL behaviour is tested against PostgreSQL. Files carry `//go:build integration`
and use `internal/platform/database/postgrestest`: one container per test binary
(started from `TestMain` through the harness), a fresh migrated database per
test. Cover:

- not-found and constraint translation
- transaction visibility and rollback
- `FOR UPDATE SKIP LOCKED` and other locking
- migration up and down

```bash
make test-integration
go test -tags integration ./internal/modules/user/...
```

Without a reachable container runtime the harness skips; set
`GOSTARTER_REQUIRE_CONTAINERS=1` to make that a failure. The
`podman-operations` skill covers the socket setup.

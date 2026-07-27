# Testing PostgreSQL Repositories

## Contents

- [Unit tests with pgxmock](#unit-tests-with-pgxmock)
- [Service tests with fakes](#service-tests-with-fakes)
- [Integration tests](#integration-tests)
- [Coverage matrix](#coverage-matrix)

## Unit tests with pgxmock

Use `github.com/pashagolub/pgxmock/v4` against the repository's query abstraction. Verify:

- parameterized SQL and arguments
- scan results and NULL behavior
- `pgx.ErrNoRows` to documented domain not-found translation
- constraint and database errors
- transaction begin/commit/rollback expectations where the repository owns them

Use regex quoting or whitespace-tolerant matching so formatting-only changes do not make
tests brittle. Always verify mock expectations.

## Service tests with fakes

Business services should receive the central domain contract or a local consumer port.
Use a small hand-written function-field fake; do not mock pgx underneath business logic.

## Integration tests

Use the repository's disposable PostgreSQL containers (Podman-backed
testcontainers harness) for integration coverage. Run real migrations and test
behavior pgxmock cannot prove:

- constraints and indexes
- `FOR UPDATE SKIP LOCKED`
- transaction visibility and rollback
- timestamp/UUID/JSON/array behavior
- migration up/down compatibility

```bash
make generate
make test-integration
# or a focused run:
go test -tags integration ./internal/platform/database/... ./internal/modules/...
```

The harness provisions its own container through the Podman API socket; no
separately started service is required.

## Coverage matrix

| Behavior | pgxmock | PostgreSQL integration |
|---|:---:|:---:|
| Query shape and arguments | ✓ | |
| Scan and error branches | ✓ | ✓ |
| Constraints and locking | | ✓ |
| Migration behavior | | ✓ |
| Business orchestration | fake port | selected end-to-end cases |

Never use a production database or add a second database/test framework without approval.

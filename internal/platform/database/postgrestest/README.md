# PostgreSQL integration-test harness

This package owns disposable PostgreSQL infrastructure for integration tests.
It is compiled only with the `integration` build tag and uses the same pinned
PostgreSQL image as local development.

Create one harness for a test, then request an isolated database:

```go
harness := postgrestest.New(t)
database := harness.Database(t)
```

`Database` applies the repository's embedded migrations. `EmptyDatabase`
creates a pristine database for migration lifecycle tests. Every call uses a
new database name, and container, pool, and database cleanup are registered
through `testing.T`.

Fixtures are explicit, ordered, parameterized statements. `ApplyFixture`
executes the complete fixture in one transaction and rolls it back on error.
Use fixed UUIDs and timestamps when a test needs deterministic data.

Run the tiers with:

```sh
make test-unit
make test-integration
make test-all
```

When Docker is unavailable, local integration tests skip with a diagnostic.
CI sets `GOSTARTER_REQUIRE_DOCKER=1`, making container startup failures fatal.
The harness never connects to a developer-owned database or binds a fixed host
port.

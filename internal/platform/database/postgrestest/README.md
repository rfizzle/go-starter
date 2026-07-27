# PostgreSQL integration-test harness

This package owns disposable PostgreSQL infrastructure for integration tests.
It is compiled only with the `integration` build tag and uses the same pinned
PostgreSQL image as local development.

One container serves a whole test binary. Install `RunMain` as the package's
`TestMain`, then request an isolated database per test:

```go
func TestMain(m *testing.M) {
	os.Exit(postgrestest.RunMain(m))
}

func TestSomething(t *testing.T) {
	database := postgrestest.Shared(t).Database(t)
	...
}
```

The container starts on the first `Shared` call and is removed before
`RunMain` returns, so a run that selects no harness-backed test starts nothing.

`Database` applies the repository's embedded migrations. `EmptyDatabase`
creates a pristine database for migration lifecycle tests. Every call uses a
new database name, and its pool and database are dropped through `testing.T`
cleanup, so tests stay isolated from each other while sharing one container.

Fixtures are explicit, ordered, parameterized statements. `ApplyFixture`
executes the complete fixture in one transaction and rolls it back on error.
Use fixed UUIDs and timestamps when a test needs deterministic data.

Run the tiers with:

```sh
make test-unit
make test-integration
make test-all
```

The harness reaches Podman through its Docker-compatible API socket; the
[environment guide](../../../../docs/development/environment.md) covers the
one-time socket setup and `DOCKER_HOST` export. When no container engine is
reachable, local integration tests skip with a diagnostic. CI sets
`GOSTARTER_REQUIRE_CONTAINERS=1`, making container startup failures fatal.
The harness never connects to a developer-owned database or binds a fixed host
port.

# Testing Strategy

Tests are executable contracts. They exist to protect useful behavior and known
invariants, not to maximize test counts or coverage percentages.

## Test Selection

Add a test only when it protects at least one of:

- An acceptance criterion or documented public contract
- A meaningful input, output, or resource boundary
- A significant failure mode owned by this code
- A previously observed regression
- A security, concurrency, persistence, or integration invariant
- Logic whose behavior would not be obvious during review

For every proposed test, identify the realistic defect it would catch. If no meaningful
defect can be named, omit the test.

Use the smallest set of cases that gives confidence in the behavior. Equivalent inputs
that prove the same invariant do not need separate tests.

Coverage is a diagnostic signal, not a completion target. An uncovered branch may reveal
an unexamined risk, but an uncovered line alone is not a reason to add a test.

## Choose the Narrowest Effective Test

Select the least expensive stable boundary that can detect the relevant failure:

| Risk | Preferred test |
| --- | --- |
| Isolated decisions, transformations, or validation | Unit test |
| Package contract observable by a consumer | Black-box package test |
| SQL, migrations, constraints, transactions, or scanning | Repository or PostgreSQL integration test |
| Generated API mapping, status, headers, or response contract | HTTP integration test |
| Cross-component user outcome | Focused end-to-end test |
| Concurrent shared-memory behavior | Deterministic test plus race detector |

Do not require unit, integration, and end-to-end tests for every change. One well-placed
integration test can be more valuable than several mocked unit tests.

## Contract Tests

Packages with multiple implementations should expose one reusable contract test suite
when the implementations must share meaningful behavior.

The suite should:

- Exercise only the documented shared contract.
- Run against every maintained implementation.
- Cover shared invariants, bounds, and error semantics.
- Leave implementation-specific behavior to implementation-specific tests.

Do not create a contract suite for a single concrete type merely to satisfy an
architectural pattern.

## Unit and Package Tests

- Test observable behavior and documented invariants rather than private call order.
- Prefer black-box `package name_test` tests when the public package contract is the
  subject. Use in-package tests when access to package-private behavior is genuinely
  needed.
- Use table-driven tests when cases share setup and assertion shape; do not force a table
  around one behavior or unrelated scenarios.
- Name tests and subtests for the behavior they protect.
- Make failures report enough context to diagnose the broken contract.
- Keep setup smaller and easier to understand than the behavior under test.

Example:

```go
func TestParseStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Status
		wantErr bool
	}{
		{name: "accepts active", input: "active", want: StatusActive},
		{name: "rejects unknown status", input: "unknown", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStatus(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

## Integration Tests

Use integration tests when the important behavior belongs to a real boundary rather than
to isolated Go logic.

### PostgreSQL

- Use the real database for migrations, constraints, query semantics, locking, and
  transaction behavior.
- Use repository mocks only when they provide a focused simulation of a consumer boundary.
- Use unique identifiers and cleanup so tests can run repeatedly.
- Tag tests that require external services with `//go:build integration`.
- Test project-owned SQL and error translation, not pgx or PostgreSQL features in
  isolation.

### HTTP

- Use `httptest.NewRequest`, `httptest.NewRecorder`, or `httptest.NewServer`.
- Assert the relevant status, headers, and response body contract.
- Cover authentication, authorization, validation, and error mapping when those behaviors
  are part of the changed contract.
- Do not call handler internals when the HTTP boundary is the behavior under test.

## Fakes, Mocks, and Helpers

- Prefer real values and lightweight fakes at true dependency boundaries.
- Model the consumer's narrow interface, not the implementation's internal calls.
- Avoid deep mock graphs and ordered expectations unless ordering is itself a public
  contract.
- Do not add a mocking, assertion, snapshot, clock, or leak-detection dependency without
  a demonstrated need and approval.
- Mark helpers with `t.Helper()`.
- Extract setup only when it reduces noise without hiding the behavior being verified.
- Prefer small, explicit fixtures over generic test-data builders with many options.

## Reliability

- Tests must be deterministic, isolated, and safe to run repeatedly.
- Do not depend on execution order, wall-clock timing, shared mutable globals, or live
  external services in the default test suite.
- Avoid arbitrary sleeps. Synchronize on observable events or use deterministic time
  control.
- Use fixed seeds when random inputs are necessary.
- Use `t.Cleanup`, `t.TempDir`, and `t.Setenv` for lifecycle and isolation.
- Use `t.Parallel()` only when the test and every fixture are concurrency-safe.
- Run race-sensitive code under the race detector.

## Regression Tests

Add a regression test when it can reproduce the defect and prevent recurrence at a stable
boundary. The test should fail for the original defect and pass for the corrected
behavior.

Do not preserve an implementation-coupled regression test when a smaller contract test
protects the same behavior.

## Tests to Avoid

Unless they protect a project-specific contract, do not test:

- Generated code
- Trivial getters, setters, constructors, or constant assignments
- Go standard-library or third-party-library behavior
- Private call order or internal implementation details
- Impossible states already excluded by types or validated boundaries
- Struct shape already enforced by the compiler
- Duplicate happy paths that add no new risk coverage

Delete obsolete, redundant, flaky, or implementation-coupled tests.

## Benchmarks and Fuzz Tests

Add a benchmark only for a performance question with a relevant workload and decision
threshold. Compare results statistically before claiming an improvement or regression.

Add a fuzz test when a parser, decoder, validator, or state transition has a broad input
space and useful invariants. Seed it with valid, empty, boundary, and previously failing
inputs.

Neither benchmarks nor fuzz tests are required for ordinary changes.

## Verification Commands

Run generation before test commands:

```bash
make generate
```

Use the smallest relevant command while developing, then the repository targets
appropriate to the changed risk:

```bash
make test
make race
make check
```

Integration tests that require local services should use the repository's integration
workflow and build tags. Do not claim a command passed unless it was run successfully.

## Review Checklist

- Each test maps to a contract, acceptance criterion, regression, boundary, or named risk.
- The selected test level is the narrowest stable level that can detect the failure.
- Cases do not duplicate the same invariant without additional value.
- Tests assert observable behavior rather than implementation structure.
- Fakes and mocks describe consumer contracts.
- Tests are deterministic and failures are diagnostic.
- No coverage target or test-count target drove the design.
- Unsupported behavior and intentionally omitted tests are documented when relevant.

For Go conventions, see [Go Standards](go-standards.md). For database behavior, see
[Database Development](../development/database.md).

# Test Naming

## Test Functions

Test functions follow `Test` + the name of what is being tested. Use underscores for subcases.

```go
func TestParseToken(t *testing.T) { ... }
func TestServer_Handle(t *testing.T) { ... }           // method test
func TestParseToken_InvalidInput(t *testing.T) { ... }  // subcase
```

## Table-Driven Tests

Table-driven test case names should be concise descriptive phrases. Lowercase prose is
common, while established acronyms may retain normal casing. Use `input` and `expected`
when those names clarify data flow.

```go
tests := []struct {
    name         string
    input        string
    expectedCode int
    expectedErr  bool
}{
    {name: "empty input", input: "", expectedCode: 400, expectedErr: true},
    {name: "valid token", input: "abc123", expectedCode: 200},
    {name: "expired token", input: "exp", expectedCode: 401, expectedErr: true},
    {name: "invalid ID", input: "???", expectedCode: 400, expectedErr: true},
}

// Bad — mixed case in test names
{name: "valid ID", ...}      // should be "valid id"
{name: "Empty Input", ...}   // should be "empty input"
```

## Test Helpers

Test helper functions that panic on failure conventionally use the `must` prefix: `mustLoadFixture()`, `mustParseURL()`.

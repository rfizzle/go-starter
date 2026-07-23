# Error Creation

## Contents

- [Plain errors](#plain-errors)
- [Sentinel errors](#sentinel-errors)
- [Typed errors](#typed-errors)
- [Messages and context](#messages-and-context)

## Plain errors

Use `errors.New` for a fixed local failure and `fmt.Errorf` when adding context or values.

```go
return errors.New("empty session id")
```

## Sentinel errors

Use a package sentinel when callers must branch on a stable condition:

```go
var ErrNotFound = errors.New("not found")
```

Wrap it without destroying identity:

```go
return fmt.Errorf("sessions.GetByID: %w", ErrNotFound)
```

Callers use `errors.Is(err, ErrNotFound)`.

## Typed errors

Use a type when callers need structured fields:

```go
type ValidationError struct {
    Field string
    Err   error
}

func (e *ValidationError) Error() string { return "invalid " + e.Field }
func (e *ValidationError) Unwrap() error { return e.Err }
```

Callers use `errors.As` to retrieve it.

## Messages and context

- Start with the failed operation: `"loading config"`, `"claiming analysis job"`.
- Keep messages lowercase and omit trailing punctuation.
- Avoid phrases such as `failed to` when the surrounding return already indicates
  failure.
- Use `%q` when string boundaries matter.
- Do not put secrets, tokens, raw SQL, or sensitive payloads in errors.
- Prefer structured attributes at the final log boundary for high-cardinality identifiers.

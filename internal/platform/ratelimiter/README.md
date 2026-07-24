# In-memory rate limiting

Package `ratelimiter` provides one bounded fixed-window limiter for protecting
process-local operations such as login attempts.

## Contract

- `MaxRequests`, `Window`, `MaxKeys`, and `CleanupInterval` must all be
  positive. `NewLimiter` rejects invalid configuration.
- Each key may make at most `MaxRequests` calls during a fixed `Window`.
- At most `MaxKeys` entries are retained. Expired entries are removed first;
  when the map is still full, the least-recently-used entry is evicted.
- `Allow`, `Reset`, and `Close` are safe for concurrent use.
- The owner must call `Close` during application shutdown. Calls to `Allow` or
  `Reset` after closure return `ErrClosed`.
- Context cancellation returns an error and never allows a request.

The limiter is deliberately in-memory. Limits are independent per process, so
deployments with multiple replicas do not receive a global aggregate limit.

## Usage

```go
limiter, err := ratelimiter.NewLimiter(ratelimiter.RateLimitConfig{
    MaxRequests:    5,
    Window:         time.Minute,
    MaxKeys:        10_000,
    CleanupInterval: time.Minute,
    KeyPrefix:      "login:",
})
if err != nil {
    return err
}
defer limiter.Close()

allowed, err := limiter.Allow(ctx, clientKey)
```

`Limiter` implements `prometheus.Collector`. Register the constructed limiter
with the application's Prometheus registry to expose:

- `rate_limiter_allowed_total`
- `rate_limiter_denied_total`
- `rate_limiter_errors_total`
- `rate_limiter_evictions_total`

These metrics have no labels, so attacker-controlled keys cannot create
unbounded metric cardinality.

`RateLimitConfig.Clock` is a deterministic test seam. Production callers
normally leave it nil to use `time.Now`.

## Non-goals

This package does not provide distributed limiting, persistent counters,
trusted-proxy client identification, or HTTP response policy. Client identity
and middleware wiring are owned by the HTTP edge and application composition.

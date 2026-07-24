// Package ratelimiter provides a bounded, process-local fixed-window rate
// limiter. A Limiter owns a cleanup goroutine that callers must stop with
// Close; its Allow, Reset, Close, and Prometheus collection methods are safe
// for concurrent use.
package ratelimiter

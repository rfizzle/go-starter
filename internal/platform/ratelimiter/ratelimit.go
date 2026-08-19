package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ErrClosed is returned when Allow or Reset is called after Close.
var ErrClosed = errors.New("rate limiter is closed")

// RateLimitConfig defines the fixed-window limit and its in-memory bounds.
type RateLimitConfig struct {
	// MaxRequests is the maximum number of requests allowed per key and window.
	MaxRequests int
	// Window is the fixed-window duration.
	Window time.Duration
	// MaxKeys is the maximum number of keys retained in memory.
	MaxKeys int
	// CleanupInterval controls how often expired keys are removed.
	CleanupInterval time.Duration
	// KeyPrefix is prepended to keys before they are stored.
	KeyPrefix string
	// Clock supplies the current time. A nil Clock uses time.Now.
	Clock func() time.Time
}

type entry struct {
	count    int
	resetAt  time.Time
	lastSeen time.Time
}

// Limiter is a bounded, in-memory fixed-window rate limiter. Its methods are
// safe for concurrent use. Close must be called to stop its cleanup goroutine.
//
// Limiter implements prometheus.Collector. Register it with the application's
// registry to expose decision, error, and eviction counters.
type Limiter struct {
	maxRequests int
	window      time.Duration
	maxKeys     int
	keyPrefix   string
	now         func() time.Time

	mu      sync.Mutex
	entries map[string]entry
	closed  bool

	done      chan struct{}
	closeOnce sync.Once
	cleanupWG sync.WaitGroup

	allowedTotal   atomic.Uint64
	deniedTotal    atomic.Uint64
	errorsTotal    atomic.Uint64
	evictionsTotal atomic.Uint64

	allowedDesc   *prometheus.Desc
	deniedDesc    *prometheus.Desc
	errorsDesc    *prometheus.Desc
	evictionsDesc *prometheus.Desc
}

// NewLimiter validates config and starts the cleanup lifecycle.
func NewLimiter(config RateLimitConfig) (*Limiter, error) {
	if config.MaxRequests <= 0 {
		return nil, fmt.Errorf("max requests must be positive: %d", config.MaxRequests)
	}
	if config.Window <= 0 {
		return nil, fmt.Errorf("window must be positive: %s", config.Window)
	}
	if config.MaxKeys <= 0 {
		return nil, fmt.Errorf("max keys must be positive: %d", config.MaxKeys)
	}
	if config.CleanupInterval <= 0 {
		return nil, fmt.Errorf("cleanup interval must be positive: %s", config.CleanupInterval)
	}

	now := config.Clock
	if now == nil {
		now = time.Now
	}

	const namespace = "rate_limiter"
	limiter := &Limiter{
		maxRequests: config.MaxRequests,
		window:      config.Window,
		maxKeys:     config.MaxKeys,
		keyPrefix:   config.KeyPrefix,
		now:         now,
		entries:     make(map[string]entry),
		done:        make(chan struct{}),
		allowedDesc: prometheus.NewDesc(
			namespace+"_allowed_total",
			"Number of requests allowed by the in-memory rate limiter.",
			nil,
			nil,
		),
		deniedDesc: prometheus.NewDesc(
			namespace+"_denied_total",
			"Number of requests denied by the in-memory rate limiter.",
			nil,
			nil,
		),
		errorsDesc: prometheus.NewDesc(
			namespace+"_errors_total",
			"Number of rate limiter operations that returned an error.",
			nil,
			nil,
		),
		evictionsDesc: prometheus.NewDesc(
			namespace+"_evictions_total",
			"Number of expired or least-recently-used keys evicted from the in-memory rate limiter.",
			nil,
			nil,
		),
	}

	limiter.cleanupWG.Add(1)
	go limiter.cleanup(config.CleanupInterval)

	return limiter, nil
}

// Allow reports whether key has remaining capacity in its current fixed
// window. Context cancellation and use after Close return errors; callers must
// not treat those errors as allowed requests.
func (l *Limiter) Allow(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		l.errorsTotal.Add(1)
		return false, err
	}

	fullKey := l.keyPrefix + key

	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		l.errorsTotal.Add(1)
		return false, ErrClosed
	}
	if err := ctx.Err(); err != nil {
		l.mu.Unlock()
		l.errorsTotal.Add(1)
		return false, err
	}
	now := l.now()

	current, exists := l.entries[fullKey]
	if !exists {
		l.evictExpiredLocked(now)
		if len(l.entries) >= l.maxKeys {
			l.evictOldestLocked()
		}

		l.entries[fullKey] = entry{
			count:    1,
			resetAt:  now.Add(l.window),
			lastSeen: now,
		}
		l.mu.Unlock()
		l.allowedTotal.Add(1)
		return true, nil
	}

	current.lastSeen = now
	if !now.Before(current.resetAt) {
		current.count = 1
		current.resetAt = now.Add(l.window)
		l.entries[fullKey] = current
		l.mu.Unlock()
		l.allowedTotal.Add(1)
		return true, nil
	}

	if current.count >= l.maxRequests {
		l.entries[fullKey] = current
		l.mu.Unlock()
		l.deniedTotal.Add(1)
		return false, nil
	}

	current.count++
	l.entries[fullKey] = current
	l.mu.Unlock()
	l.allowedTotal.Add(1)
	return true, nil
}

// Reset removes the current window for key.
func (l *Limiter) Reset(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		l.errorsTotal.Add(1)
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		l.errorsTotal.Add(1)
		return ErrClosed
	}
	if err := ctx.Err(); err != nil {
		l.errorsTotal.Add(1)
		return err
	}

	delete(l.entries, l.keyPrefix+key)
	return nil
}

// Close stops background cleanup and prevents further Allow or Reset calls.
// It is safe to call concurrently and more than once.
func (l *Limiter) Close() {
	l.closeOnce.Do(func() {
		l.mu.Lock()
		l.closed = true
		l.mu.Unlock()

		close(l.done)
		l.cleanupWG.Wait()
	})
}

func (l *Limiter) cleanup(interval time.Duration) {
	defer l.cleanupWG.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := l.now()
			l.evictExpiredLocked(now)
			l.mu.Unlock()
		case <-l.done:
			return
		}
	}
}

func (l *Limiter) evictExpiredLocked(now time.Time) {
	for key, current := range l.entries {
		if !now.Before(current.resetAt) {
			delete(l.entries, key)
			l.evictionsTotal.Add(1)
		}
	}
}

func (l *Limiter) evictOldestLocked() {
	var oldestKey string
	var oldestTime time.Time
	found := false
	for key, current := range l.entries {
		if !found || current.lastSeen.Before(oldestTime) {
			oldestKey = key
			oldestTime = current.lastSeen
			found = true
		}
	}
	if found {
		delete(l.entries, oldestKey)
		l.evictionsTotal.Add(1)
	}
}

// Describe implements prometheus.Collector.
func (l *Limiter) Describe(ch chan<- *prometheus.Desc) {
	ch <- l.allowedDesc
	ch <- l.deniedDesc
	ch <- l.errorsDesc
	ch <- l.evictionsDesc
}

// Collect implements prometheus.Collector.
func (l *Limiter) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(l.allowedDesc, prometheus.CounterValue, float64(l.allowedTotal.Load()))
	ch <- prometheus.MustNewConstMetric(l.deniedDesc, prometheus.CounterValue, float64(l.deniedTotal.Load()))
	ch <- prometheus.MustNewConstMetric(l.errorsDesc, prometheus.CounterValue, float64(l.errorsTotal.Load()))
	ch <- prometheus.MustNewConstMetric(l.evictionsDesc, prometheus.CounterValue, float64(l.evictionsTotal.Load()))
}

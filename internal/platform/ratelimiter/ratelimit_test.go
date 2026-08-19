package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func validConfig(clock *fakeClock) RateLimitConfig {
	return RateLimitConfig{
		MaxRequests:     2,
		Window:          time.Minute,
		MaxKeys:         4,
		CleanupInterval: time.Hour,
		KeyPrefix:       "test:",
		Clock:           clock.Now,
	}
}

func newTestLimiter(t *testing.T, config RateLimitConfig) *Limiter {
	t.Helper()

	limiter, err := NewLimiter(config)
	require.NoError(t, err)
	t.Cleanup(limiter.Close)
	return limiter
}

func TestNewLimiter_ValidatesConfig(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1, 0)}

	tests := []struct {
		name   string
		mutate func(*RateLimitConfig)
		want   string
	}{
		{
			name:   "zero max requests",
			mutate: func(config *RateLimitConfig) { config.MaxRequests = 0 },
			want:   "max requests must be positive",
		},
		{
			name:   "negative max requests",
			mutate: func(config *RateLimitConfig) { config.MaxRequests = -1 },
			want:   "max requests must be positive",
		},
		{
			name:   "zero window",
			mutate: func(config *RateLimitConfig) { config.Window = 0 },
			want:   "window must be positive",
		},
		{
			name:   "negative window",
			mutate: func(config *RateLimitConfig) { config.Window = -time.Second },
			want:   "window must be positive",
		},
		{
			name:   "zero max keys",
			mutate: func(config *RateLimitConfig) { config.MaxKeys = 0 },
			want:   "max keys must be positive",
		},
		{
			name:   "negative max keys",
			mutate: func(config *RateLimitConfig) { config.MaxKeys = -1 },
			want:   "max keys must be positive",
		},
		{
			name:   "zero cleanup interval",
			mutate: func(config *RateLimitConfig) { config.CleanupInterval = 0 },
			want:   "cleanup interval must be positive",
		},
		{
			name:   "negative cleanup interval",
			mutate: func(config *RateLimitConfig) { config.CleanupInterval = -time.Second },
			want:   "cleanup interval must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := validConfig(clock)
			tt.mutate(&config)

			limiter, err := NewLimiter(config)
			require.ErrorContains(t, err, tt.want)
			assert.Nil(t, limiter)
		})
	}
}

func TestLimiter_AllowResetAndWindowExpiry(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1, 0)}
	limiter := newTestLimiter(t, validConfig(clock))
	ctx := context.Background()

	for request := 1; request <= 2; request++ {
		allowed, err := limiter.Allow(ctx, "client")
		require.NoError(t, err)
		assert.True(t, allowed, "request %d should be allowed", request)
	}

	allowed, err := limiter.Allow(ctx, "client")
	require.NoError(t, err)
	assert.False(t, allowed)

	require.NoError(t, limiter.Reset(ctx, "client"))
	allowed, err = limiter.Allow(ctx, "client")
	require.NoError(t, err)
	assert.True(t, allowed)

	clock.Advance(time.Minute)
	allowed, err = limiter.Allow(ctx, "client")
	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestLimiter_BoundsHighCardinalityAndExpiresKeys(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1, 0)}
	config := validConfig(clock)
	config.MaxKeys = 32
	limiter := newTestLimiter(t, config)
	ctx := context.Background()

	for i := 0; i < 10_000; i++ {
		allowed, err := limiter.Allow(ctx, fmt.Sprintf("client-%d", i))
		require.NoError(t, err)
		require.True(t, allowed)

		limiter.mu.Lock()
		entryCount := len(limiter.entries)
		limiter.mu.Unlock()
		if entryCount > config.MaxKeys {
			t.Fatalf("entry count = %d, want at most %d", entryCount, config.MaxKeys)
		}
	}

	clock.Advance(config.Window)
	allowed, err := limiter.Allow(ctx, "after-expiry")
	require.NoError(t, err)
	assert.True(t, allowed)

	limiter.mu.Lock()
	entryCount := len(limiter.entries)
	limiter.mu.Unlock()
	assert.Equal(t, 1, entryCount)
}

func TestLimiter_Metrics(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1, 0)}
	config := validConfig(clock)
	config.MaxRequests = 1
	config.MaxKeys = 1
	limiter := newTestLimiter(t, config)

	allowed, err := limiter.Allow(context.Background(), "first")
	require.NoError(t, err)
	require.True(t, allowed)

	allowed, err = limiter.Allow(context.Background(), "first")
	require.NoError(t, err)
	require.False(t, allowed)

	allowed, err = limiter.Allow(context.Background(), "second")
	require.NoError(t, err)
	require.True(t, allowed)

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	allowed, err = limiter.Allow(cancelled, "third")
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, allowed)

	registry := prometheus.NewRegistry()
	require.NoError(t, registry.Register(limiter))
	families, err := registry.Gather()
	require.NoError(t, err)

	values := make(map[string]float64, len(families))
	for _, family := range families {
		require.Len(t, family.GetMetric(), 1)
		values[family.GetName()] = family.GetMetric()[0].GetCounter().GetValue()
	}

	assert.Equal(t, float64(2), values["rate_limiter_allowed_total"])
	assert.Equal(t, float64(1), values["rate_limiter_denied_total"])
	assert.Equal(t, float64(1), values["rate_limiter_errors_total"])
	assert.Equal(t, float64(1), values["rate_limiter_evictions_total"])
}

func TestLimiter_ConcurrentAllowResetAndClose(t *testing.T) {
	clock := &fakeClock{now: time.Unix(1, 0)}
	config := validConfig(clock)
	config.MaxRequests = 100
	config.MaxKeys = 16
	limiter := newTestLimiter(t, config)

	const workers = 32
	start := make(chan struct{})
	errs := make(chan error, workers)
	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start

			for operation := 0; operation < 100; operation++ {
				key := fmt.Sprintf("client-%d", id%config.MaxKeys)
				var err error
				if operation%3 == 0 {
					err = limiter.Reset(context.Background(), key)
				} else {
					_, err = limiter.Allow(context.Background(), key)
				}
				if err != nil && !errors.Is(err, ErrClosed) {
					errs <- err
					return
				}
			}
		}(worker)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		limiter.Close()
		limiter.Close()
	}()

	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent operation returned unexpected error: %v", err)
	}

	allowed, err := limiter.Allow(context.Background(), "after-close")
	assert.False(t, allowed)
	assert.ErrorIs(t, err, ErrClosed)
	assert.ErrorIs(t, limiter.Reset(context.Background(), "after-close"), ErrClosed)
}

var _ prometheus.Collector = (*Limiter)(nil)

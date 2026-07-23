package ratelimiter

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryRateLimiter(t *testing.T) {
	config := RateLimitConfig{
		MaxRequests: 3,
		Window:      100 * time.Millisecond,
		KeyPrefix:   "test:",
	}

	limiter := NewMemoryRateLimiter(config)
	ctx := context.Background()

	t.Run("allows requests within limit", func(t *testing.T) {
		key := "user1"

		// First 3 requests should be allowed
		for i := 0; i < 3; i++ {
			allowed, err := limiter.Allow(ctx, key)
			assert.NoError(t, err)
			assert.True(t, allowed, "request %d should be allowed", i+1)
		}

		// 4th request should be denied
		allowed, err := limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.False(t, allowed, "4th request should be denied")
	})

	t.Run("resets after window expires", func(t *testing.T) {
		key := "user2"

		// Use up the limit
		for i := 0; i < 3; i++ {
			allowed, err := limiter.Allow(ctx, key)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		// Should be denied
		allowed, err := limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.False(t, allowed)

		// Wait for window to expire
		time.Sleep(110 * time.Millisecond)

		// Should be allowed again
		allowed, err = limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("different keys are independent", func(t *testing.T) {
		key1 := "user3"
		key2 := "user4"

		// Use up limit for key1
		for i := 0; i < 3; i++ {
			allowed, err := limiter.Allow(ctx, key1)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		// key1 should be denied
		allowed, err := limiter.Allow(ctx, key1)
		assert.NoError(t, err)
		assert.False(t, allowed)

		// key2 should still be allowed
		allowed, err = limiter.Allow(ctx, key2)
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("reset clears rate limit", func(t *testing.T) {
		key := "user5"

		// Use up the limit
		for i := 0; i < 3; i++ {
			allowed, err := limiter.Allow(ctx, key)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		// Should be denied
		allowed, err := limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.False(t, allowed)

		// Reset the limit
		err = limiter.Reset(ctx, key)
		assert.NoError(t, err)

		// Should be allowed again
		allowed, err = limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.True(t, allowed)
	})
}

func TestMemoryRateLimiter_GetCount(t *testing.T) {
	config := RateLimitConfig{
		MaxRequests: 5,
		Window:      time.Hour,
		KeyPrefix:   "test:",
	}

	limiter := NewMemoryRateLimiter(config)
	ctx := context.Background()
	key := "user1"

	// Initially should be 0
	count, err := limiter.GetCount(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// Make 3 requests
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.True(t, allowed)
	}

	// Count should be 3
	count, err = limiter.GetCount(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestMemoryRateLimiter_Concurrent(t *testing.T) {
	config := RateLimitConfig{
		MaxRequests: 10,
		Window:      time.Second,
		KeyPrefix:   "test:",
	}

	limiter := NewMemoryRateLimiter(config)
	ctx := context.Background()
	key := "concurrent_user"

	var wg sync.WaitGroup
	var allowedCount int32
	var deniedCount int32
	var mu sync.Mutex

	// Launch 20 concurrent requests
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, err := limiter.Allow(ctx, key)
			assert.NoError(t, err)

			mu.Lock()
			if allowed {
				allowedCount++
			} else {
				deniedCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Should have exactly 10 allowed and 10 denied
	assert.Equal(t, int32(10), allowedCount)
	assert.Equal(t, int32(10), deniedCount)
}

func TestSlidingWindowRateLimiter(t *testing.T) {
	config := RateLimitConfig{
		MaxRequests: 3,
		Window:      200 * time.Millisecond,
		KeyPrefix:   "sliding:",
	}

	limiter := NewSlidingWindowRateLimiter(config)
	ctx := context.Background()

	t.Run("allows requests within sliding window", func(t *testing.T) {
		key := "sliding_user1"

		// Make 3 requests - should all be allowed
		for i := 0; i < 3; i++ {
			allowed, err := limiter.Allow(ctx, key)
			assert.NoError(t, err)
			assert.True(t, allowed, "request %d should be allowed", i+1)
		}

		// 4th request should be denied
		allowed, err := limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.False(t, allowed, "4th request should be denied")

		// Wait for window to slide partially
		time.Sleep(100 * time.Millisecond)

		// Still within window, should be denied
		allowed, err = limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.False(t, allowed)

		// Wait for full window to pass
		time.Sleep(120 * time.Millisecond)

		// Should be allowed again
		allowed, err = limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.True(t, allowed)
	})
}

func TestTokenBucketRateLimiter(t *testing.T) {
	config := RateLimitConfig{
		MaxRequests: 5,                      // 5 tokens capacity
		Window:      500 * time.Millisecond, // Refill 1 token every 100ms
		KeyPrefix:   "bucket:",
	}

	limiter := NewTokenBucketRateLimiter(config)
	ctx := context.Background()

	t.Run("initial burst allowed", func(t *testing.T) {
		key := "bucket_user1"

		// Should allow initial burst up to capacity
		for i := 0; i < 5; i++ {
			allowed, err := limiter.Allow(ctx, key)
			assert.NoError(t, err)
			assert.True(t, allowed, "request %d should be allowed", i+1)
		}

		// 6th request should be denied (bucket empty)
		allowed, err := limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.False(t, allowed, "6th request should be denied")
	})

	t.Run("tokens refill over time", func(t *testing.T) {
		key := "bucket_user2"

		// Use up all tokens
		for i := 0; i < 5; i++ {
			allowed, err := limiter.Allow(ctx, key)
			assert.NoError(t, err)
			assert.True(t, allowed)
		}

		// Should be denied
		allowed, err := limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.False(t, allowed)

		// Wait for some tokens to refill
		time.Sleep(150 * time.Millisecond) // Should refill 1 token

		// Should allow 1 request
		allowed, err = limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.True(t, allowed)

		// Should be denied again
		allowed, err = limiter.Allow(ctx, key)
		assert.NoError(t, err)
		assert.False(t, allowed)
	})
}

func TestRateLimitConfig_Defaults(t *testing.T) {
	config := DefaultRateLimitConfig()

	assert.Equal(t, 100, config.MaxRequests)
	assert.Equal(t, time.Hour, config.Window)
	assert.Equal(t, "rate_limit:", config.KeyPrefix)
}

func TestRateLimitEntry(t *testing.T) {
	now := time.Now()

	t.Run("IsExpired", func(t *testing.T) {
		// Not expired
		entry := &RateLimitEntry{
			ExpiresAt: now.Add(time.Hour),
		}
		assert.False(t, entry.IsExpired())

		// Expired
		entry.ExpiresAt = now.Add(-time.Hour)
		assert.True(t, entry.IsExpired())
	})

	t.Run("ShouldReset", func(t *testing.T) {
		// Should not reset
		entry := &RateLimitEntry{
			Window: now.Add(time.Hour),
		}
		assert.False(t, entry.ShouldReset())

		// Should reset
		entry.Window = now.Add(-time.Hour)
		assert.True(t, entry.ShouldReset())
	})
}

func TestRedisRateLimiter_Interface(t *testing.T) {
	// Just test that RedisRateLimiter implements the interface
	config := DefaultRateLimitConfig()
	limiter := NewRedisRateLimiter(config)

	var _ RateLimiter = limiter

	ctx := context.Background()

	// These will use placeholder implementation
	allowed, err := limiter.Allow(ctx, "test")
	assert.NoError(t, err)
	assert.True(t, allowed) // Placeholder always returns true

	err = limiter.Reset(ctx, "test")
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkMemoryRateLimiter_Allow(b *testing.B) {
	config := RateLimitConfig{
		MaxRequests: 1000,
		Window:      time.Hour,
		KeyPrefix:   "bench:",
	}

	limiter := NewMemoryRateLimiter(config)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "user" + string(rune(i%100)) // Use 100 different keys
			_, _ = limiter.Allow(ctx, key)
			i++
		}
	})
}

func BenchmarkSlidingWindowRateLimiter_Allow(b *testing.B) {
	config := RateLimitConfig{
		MaxRequests: 1000,
		Window:      time.Hour,
		KeyPrefix:   "bench:",
	}

	limiter := NewSlidingWindowRateLimiter(config)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "user" + string(rune(i%100))
			_, _ = limiter.Allow(ctx, key)
			i++
		}
	})
}

func BenchmarkTokenBucketRateLimiter_Allow(b *testing.B) {
	config := RateLimitConfig{
		MaxRequests: 1000,
		Window:      time.Hour,
		KeyPrefix:   "bench:",
	}

	limiter := NewTokenBucketRateLimiter(config)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "user" + string(rune(i%100))
			_, _ = limiter.Allow(ctx, key)
			i++
		}
	})
}

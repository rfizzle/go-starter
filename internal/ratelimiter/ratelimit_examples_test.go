package ratelimiter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Example usage tests showing how to use different rate limiters directly

func TestExampleMemoryRateLimiter_basicUsage(t *testing.T) {
	// Configure rate limiter: 5 requests per minute
	config := RateLimitConfig{
		MaxRequests: 5,
		Window:      time.Minute,
		KeyPrefix:   "api:",
	}

	rateLimiter := NewMemoryRateLimiter(config)
	ctx := context.Background()
	clientKey := "192.168.1.1"

	// Simulate 5 requests from same IP - should all succeed
	for i := 0; i < 5; i++ {
		allowed, err := rateLimiter.Allow(ctx, clientKey)
		assert.NoError(t, err)
		assert.True(t, allowed, "Request %d should succeed", i+1)
	}

	// 6th request should be rate limited
	allowed, err := rateLimiter.Allow(ctx, clientKey)
	assert.NoError(t, err)
	assert.False(t, allowed, "6th request should be denied")
}

func TestExampleSlidingWindowRateLimiter_burst(t *testing.T) {
	// Configure sliding window: 10 requests per 30 seconds
	config := RateLimitConfig{
		MaxRequests: 10,
		Window:      30 * time.Second,
		KeyPrefix:   "burst:",
	}

	rateLimiter := NewSlidingWindowRateLimiter(config)
	ctx := context.Background()
	clientKey := "client123"

	// Allow initial burst of 10 requests
	for i := 0; i < 10; i++ {
		allowed, err := rateLimiter.Allow(ctx, clientKey)
		assert.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 11th request should be denied
	allowed, err := rateLimiter.Allow(ctx, clientKey)
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Different client should still be allowed
	allowed, err = rateLimiter.Allow(ctx, "different_client")
	assert.NoError(t, err)
	assert.True(t, allowed)
}

func TestExampleTokenBucketRateLimiter_steadyRate(t *testing.T) {
	// Configure token bucket: 60 requests per hour (1 per minute)
	config := RateLimitConfig{
		MaxRequests: 10,               // Bucket capacity
		Window:      10 * time.Second, // Refill 1 token every second
		KeyPrefix:   "steady:",
	}

	rateLimiter := NewTokenBucketRateLimiter(config)
	ctx := context.Background()
	clientKey := "steady_client"

	// Initial burst should be allowed (up to capacity)
	for i := 0; i < 10; i++ {
		allowed, err := rateLimiter.Allow(ctx, clientKey)
		assert.NoError(t, err)
		assert.True(t, allowed, "Initial burst request %d should be allowed", i+1)
	}

	// Next request should be denied (bucket empty)
	allowed, err := rateLimiter.Allow(ctx, clientKey)
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Note: In real usage, you'd wait for tokens to refill
	// or use a longer time window for testing
}

func TestExampleRateLimiter_monitoring(t *testing.T) {
	config := RateLimitConfig{
		MaxRequests: 10,
		Window:      time.Minute,
		KeyPrefix:   "monitor:",
	}

	rateLimiter := NewMemoryRateLimiter(config)
	ctx := context.Background()
	clientKey := "monitored_client"

	// Make some requests
	for i := 0; i < 7; i++ {
		allowed, err := rateLimiter.Allow(ctx, clientKey)
		assert.NoError(t, err)
		assert.True(t, allowed)
	}

	// Check current usage
	count, err := rateLimiter.GetCount(ctx, clientKey)
	assert.NoError(t, err)
	assert.Equal(t, 7, count)

	// You could use this for monitoring/alerting
	if count > 5 {
		t.Logf("Client %s is approaching rate limit: %d/%d", clientKey, count, config.MaxRequests)
	}

	// Reset if needed (e.g., for admin override)
	err = rateLimiter.Reset(ctx, clientKey)
	assert.NoError(t, err)

	// Should be back to 0
	count, err = rateLimiter.GetCount(ctx, clientKey)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

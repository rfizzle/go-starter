package ratelimiter

import (
	"context"
	"sync"
	"time"
)

// RateLimiter interface for rate limiting authentication attempts
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
	Reset(ctx context.Context, key string) error
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// MaxRequests is the maximum number of requests allowed
	MaxRequests int
	// Window is the time window for the rate limit
	Window time.Duration
	// KeyPrefix is prepended to all keys
	KeyPrefix string
}

// DefaultRateLimitConfig returns sensible rate limiting defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRequests: 100,
		Window:      time.Hour,
		KeyPrefix:   "rate_limit:",
	}
}

// RateLimitEntry represents a single rate limit entry
type RateLimitEntry struct {
	Count     int
	Window    time.Time
	ExpiresAt time.Time
}

// IsExpired checks if the rate limit entry has expired
func (e *RateLimitEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// ShouldReset checks if the window has passed and count should reset
func (e *RateLimitEntry) ShouldReset() bool {
	return time.Now().After(e.Window)
}

// MemoryRateLimiter implements rate limiting using in-memory storage
type MemoryRateLimiter struct {
	config  RateLimitConfig
	entries map[string]*RateLimitEntry
	mutex   sync.RWMutex
}

// NewMemoryRateLimiter creates a new in-memory rate limiter
func NewMemoryRateLimiter(config RateLimitConfig) *MemoryRateLimiter {
	limiter := &MemoryRateLimiter{
		config:  config,
		entries: make(map[string]*RateLimitEntry),
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a request is allowed for the given key
func (m *MemoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	fullKey := m.config.KeyPrefix + key

	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	entry, exists := m.entries[fullKey]

	if !exists || entry.IsExpired() {
		// Create new entry
		m.entries[fullKey] = &RateLimitEntry{
			Count:     1,
			Window:    now.Add(m.config.Window),
			ExpiresAt: now.Add(m.config.Window * 2), // Keep for double the window for accurate tracking
		}
		return true, nil
	}

	if entry.ShouldReset() {
		// Reset the window
		entry.Count = 1
		entry.Window = now.Add(m.config.Window)
		entry.ExpiresAt = now.Add(m.config.Window * 2)
		return true, nil
	}

	if entry.Count >= m.config.MaxRequests {
		return false, nil
	}

	entry.Count++
	return true, nil
}

// Reset clears the rate limit for the given key
func (m *MemoryRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := m.config.KeyPrefix + key

	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.entries, fullKey)
	return nil
}

// GetCount returns the current count for debugging/monitoring
func (m *MemoryRateLimiter) GetCount(ctx context.Context, key string) (int, error) {
	fullKey := m.config.KeyPrefix + key

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	entry, exists := m.entries[fullKey]
	if !exists || entry.IsExpired() {
		return 0, nil
	}

	if entry.ShouldReset() {
		return 0, nil
	}

	return entry.Count, nil
}

// cleanup periodically removes expired entries
func (m *MemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(m.config.Window)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		for key, entry := range m.entries {
			if entry.IsExpired() {
				delete(m.entries, key)
			}
		}
		m.mutex.Unlock()
	}
}

// RedisRateLimiter implements rate limiting using Redis
// This is a skeleton implementation - you'd need to add actual Redis client
type RedisRateLimiter struct {
	config RateLimitConfig
	// redis  redis.Client // You would inject your Redis client here
}

// NewRedisRateLimiter creates a new Redis-backed rate limiter
func NewRedisRateLimiter(config RateLimitConfig /* , redisClient redis.Client */) *RedisRateLimiter {
	return &RedisRateLimiter{
		config: config,
		// redis:  redisClient,
	}
}

// Allow checks if a request is allowed using Redis
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	_ = key // Suppress unused warning for placeholder implementation

	// Example Redis implementation (pseudo-code):
	// 1. Use Redis INCR to increment counter
	// 2. Set expiration if it's the first increment
	// 3. Check if count exceeds limit

	// Lua script for atomic increment and check:
	/*
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])

		local current = redis.call("GET", key)
		if current == false then
			redis.call("SET", key, 1)
			redis.call("EXPIRE", key, window)
			return 1
		end

		current = tonumber(current)
		if current < limit then
			return redis.call("INCR", key)
		else
			return -1
		end
	*/

	// For now, return true (implement with your Redis client)
	return true, nil
}

// Reset clears the rate limit for the given key using Redis
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	_ = key // Suppress unused warning for placeholder implementation

	// Example: r.redis.Del(ctx, fullKey)
	return nil
}

// SlidingWindowRateLimiter implements a sliding window rate limiter
type SlidingWindowRateLimiter struct {
	config  RateLimitConfig
	windows map[string][]time.Time
	mutex   sync.RWMutex
}

// NewSlidingWindowRateLimiter creates a sliding window rate limiter
func NewSlidingWindowRateLimiter(config RateLimitConfig) *SlidingWindowRateLimiter {
	limiter := &SlidingWindowRateLimiter{
		config:  config,
		windows: make(map[string][]time.Time),
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if a request is allowed using sliding window
func (s *SlidingWindowRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	fullKey := s.config.KeyPrefix + key

	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-s.config.Window)

	// Get existing timestamps for this key
	timestamps, exists := s.windows[fullKey]
	if !exists {
		timestamps = make([]time.Time, 0, s.config.MaxRequests+1)
	}

	// Remove timestamps outside the window
	validTimestamps := make([]time.Time, 0, len(timestamps))
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if we can add another request
	if len(validTimestamps) >= s.config.MaxRequests {
		s.windows[fullKey] = validTimestamps
		return false, nil
	}

	// Add current timestamp
	validTimestamps = append(validTimestamps, now)
	s.windows[fullKey] = validTimestamps

	return true, nil
}

// Reset clears the rate limit for the given key
func (s *SlidingWindowRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := s.config.KeyPrefix + key

	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.windows, fullKey)
	return nil
}

// cleanup periodically removes old entries
func (s *SlidingWindowRateLimiter) cleanup() {
	ticker := time.NewTicker(s.config.Window)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		windowStart := time.Now().Add(-s.config.Window * 2) // Keep some buffer

		for key, timestamps := range s.windows {
			validTimestamps := make([]time.Time, 0, len(timestamps))
			for _, ts := range timestamps {
				if ts.After(windowStart) {
					validTimestamps = append(validTimestamps, ts)
				}
			}

			if len(validTimestamps) == 0 {
				delete(s.windows, key)
			} else {
				s.windows[key] = validTimestamps
			}
		}
		s.mutex.Unlock()
	}
}

// TokenBucketRateLimiter implements a token bucket rate limiter
type TokenBucketRateLimiter struct {
	config  RateLimitConfig
	buckets map[string]*TokenBucket
	mutex   sync.RWMutex
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens     int
	capacity   int
	refillRate time.Duration
	lastRefill time.Time
	mutex      sync.Mutex
}

// NewTokenBucketRateLimiter creates a token bucket rate limiter
func NewTokenBucketRateLimiter(config RateLimitConfig) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		config:  config,
		buckets: make(map[string]*TokenBucket),
	}
}

// Allow checks if a request is allowed using token bucket algorithm
func (t *TokenBucketRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	fullKey := t.config.KeyPrefix + key

	t.mutex.RLock()
	bucket, exists := t.buckets[fullKey]
	t.mutex.RUnlock()

	if !exists {
		t.mutex.Lock()
		// Double-check after acquiring write lock
		bucket, exists = t.buckets[fullKey]
		if !exists {
			bucket = &TokenBucket{
				tokens:     t.config.MaxRequests - 1, // Take one token immediately
				capacity:   t.config.MaxRequests,
				refillRate: t.config.Window / time.Duration(t.config.MaxRequests),
				lastRefill: time.Now(),
			}
			t.buckets[fullKey] = bucket
		}
		t.mutex.Unlock()

		if !exists {
			return true, nil // First request is always allowed
		}
	}

	return bucket.takeToken(), nil
}

// takeToken attempts to take a token from the bucket
func (b *TokenBucket) takeToken() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	now := time.Now()
	timePassed := now.Sub(b.lastRefill)

	// Calculate tokens to add based on time passed
	tokensToAdd := int(timePassed / b.refillRate)
	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
		b.lastRefill = now
	}

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// Reset clears the rate limit for the given key
func (t *TokenBucketRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := t.config.KeyPrefix + key

	t.mutex.Lock()
	defer t.mutex.Unlock()

	delete(t.buckets, fullKey)
	return nil
}

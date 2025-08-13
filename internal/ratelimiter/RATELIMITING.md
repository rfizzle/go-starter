# Rate Limiting Implementation Guide

This document describes the rate limiting implementations.

## Overview

The ratelimiter package provides multiple rate limiting strategies with a common interface, allowing you to choose the best approach for your application's needs.

## Interface

All rate limiters implement the `RateLimiter` interface:

```go
type RateLimiter interface {
    Allow(ctx context.Context, key string) (bool, error)
    Reset(ctx context.Context, key string) error
}
```

## Available Implementations

### 1. Memory Rate Limiter

**Best for**: Single-instance applications, development, or when you don't need distributed rate limiting.

```go
config := RateLimitConfig{
    MaxRequests: 100,
    Window:      time.Hour,
    KeyPrefix:   "api:",
}
rateLimiter := NewMemoryRateLimiter(config)
```

**Features**:
- Fast, in-memory storage
- Automatic cleanup of expired entries
- Thread-safe with mutex protection
- `GetCount()` method for monitoring
- Fixed window algorithm

**Pros**:
- Very fast (no network calls)
- Simple setup
- Built-in monitoring

**Cons**:
- Not distributed (doesn't work across multiple instances)
- Memory usage grows with unique keys
- Rate limit resets abruptly at window boundaries

### 2. Sliding Window Rate Limiter

**Best for**: Applications needing smoother rate limiting without abrupt resets.

```go
config := RateLimitConfig{
    MaxRequests: 60,
    Window:      time.Minute,
    KeyPrefix:   "smooth:",
}
rateLimiter := NewSlidingWindowRateLimiter(config)
```

**Features**:
- Sliding window algorithm
- Smoother rate limiting
- Automatic cleanup
- More accurate rate limiting

**Pros**:
- No abrupt resets at window boundaries
- More accurate rate limiting
- Better user experience

**Cons**:
- Higher memory usage (stores timestamps)
- More CPU intensive
- Not distributed

### 3. Token Bucket Rate Limiter

**Best for**: Applications needing burst tolerance with steady-state rate limiting.

```go
config := RateLimitConfig{
    MaxRequests: 10,                  // Bucket capacity
    Window:      10 * time.Second,    // Time to refill bucket
    KeyPrefix:   "burst:",
}
rateLimiter := NewTokenBucketRateLimiter(config)
```

**Features**:
- Token bucket algorithm
- Allows initial bursts
- Steady refill rate
- Natural burst handling

**Pros**:
- Handles traffic bursts gracefully
- More flexible than fixed windows
- Good for APIs with occasional spikes

**Cons**:
- More complex to configure
- Not distributed
- Requires understanding of token refill rates

### 4. Redis Rate Limiter (Skeleton)

**Best for**: Distributed applications needing shared rate limiting across multiple instances.

```go
config := RateLimitConfig{
    MaxRequests: 1000,
    Window:      time.Hour,
    KeyPrefix:   "distributed:",
}
// You need to implement with your Redis client
rateLimiter := NewRedisRateLimiter(config)
```

**Features** (when implemented):
- Distributed rate limiting
- Persistence across restarts
- Atomic operations with Lua scripts
- Scales horizontally

**Implementation needed**:
- Add Redis client dependency
- Implement Allow() and Reset() methods
- Use Lua scripts for atomic operations

## Usage Examples

### Basic Setup with Memory Limiter

```go
// Create rate limiter
config := RateLimitConfig{
    MaxRequests: 100,
    Window:      time.Hour,
    KeyPrefix:   "api:",
}
rateLimiter := NewMemoryRateLimiter(config)

// Create auth with rate limiting middleware
validator := &YourValidator{}
auth, err := NewAuth(validator, WithLogger(logger))
if err != nil {
    log.Fatal(err)
}

// Apply middleware
router.Use(auth.RateLimitMiddleware(rateLimiter))
```

### Different Limits for Different Endpoints

```go
// API endpoints: 1000 requests/hour
apiLimiter := NewMemoryRateLimiter(RateLimitConfig{
    MaxRequests: 1000,
    Window:      time.Hour,
    KeyPrefix:   "api:",
})

// Auth endpoints: 10 requests/minute  
authLimiter := NewMemoryRateLimiter(RateLimitConfig{
    MaxRequests: 10,
    Window:      time.Minute,
    KeyPrefix:   "auth:",
})

// Apply to different route groups
apiRoutes.Use(auth.RateLimitMiddleware(apiLimiter))
authRoutes.Use(auth.RateLimitMiddleware(authLimiter))
```

### Custom Key Extraction

```go
// Rate limit by user ID instead of IP
userRateLimitMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID, err := auth.subjectFromContext(r.Context())
        if err != nil {
            userID = getClientIP(r) // Fallback to IP
        }
        
        allowed, err := rateLimiter.Allow(r.Context(), userID)
        if err != nil {
            http.Error(w, "Internal Server Error", 500)
            return
        }
        
        if !allowed {
            http.Error(w, "Too Many Requests", 429)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### Monitoring and Alerting

```go
// Monitor rate limit usage
if memoryLimiter, ok := rateLimiter.(*MemoryRateLimiter); ok {
    count, _ := memoryLimiter.GetCount(ctx, clientKey)
    if count > config.MaxRequests * 0.8 { // 80% threshold
        logger.Warn("Client approaching rate limit", 
            zap.String("client", clientKey),
            zap.Int("current", count),
            zap.Int("limit", config.MaxRequests))
    }
}
```

### Admin Override

```go
// Reset rate limit for specific client
func resetRateLimit(rateLimiter RateLimiter, clientKey string) error {
    return rateLimiter.Reset(context.Background(), clientKey)
}
```

## Configuration Guidelines

### Memory Limiter
```go
// High-traffic API
RateLimitConfig{
    MaxRequests: 10000,
    Window:      time.Hour,
    KeyPrefix:   "api:",
}

// Authentication endpoints
RateLimitConfig{
    MaxRequests: 5,
    Window:      time.Minute,
    KeyPrefix:   "auth:",
}
```

### Sliding Window Limiter
```go
// Smooth rate limiting for user actions
RateLimitConfig{
    MaxRequests: 60,        // 1 request per second average
    Window:      time.Minute,
    KeyPrefix:   "user:",
}
```

### Token Bucket Limiter
```go
// Allow bursts but limit sustained rate
RateLimitConfig{
    MaxRequests: 20,                // 20 token capacity
    Window:      20 * time.Second,  // 1 token per second refill
    KeyPrefix:   "burst:",
}
```

## Performance Characteristics

| Implementation | Memory Usage | CPU Usage | Accuracy  | Distribution |
|----------------|--------------|-----------|-----------|--------------|
| Memory         | Low          | Very Low  | Good      | Single Node  |
| Sliding Window | High         | Medium    | Excellent | Single Node  |
| Token Bucket   | Medium       | Low       | Good      | Single Node  |
| Redis          | Low (local)  | Low       | Good      | Distributed  |

## Best Practices

1. **Choose the right algorithm**:
   - Memory limiter for simple cases
   - Sliding window for smooth limiting
   - Token bucket for burst tolerance
   - Redis for distributed systems

2. **Set appropriate key prefixes**:
   ```go
   "api:"     // API endpoints
   "auth:"    // Authentication
   "user:"    // Per-user limits
   "admin:"   // Admin operations
   ```

3. **Configure reasonable limits**:
   - Start conservative and increase based on monitoring
   - Different limits for different endpoint types
   - Consider your backend capacity

4. **Monitor rate limiting**:
   - Track how often limits are hit
   - Monitor for legitimate traffic patterns
   - Set up alerting for unusual patterns

5. **Handle errors gracefully**:
   ```go
   allowed, err := rateLimiter.Allow(ctx, key)
   if err != nil {
       // Log error but don't block requests
       logger.Error("Rate limiter error", zap.Error(err))
       allowed = true // Fail open
   }
   ```

6. **Test your rate limits**:
   - Unit tests for edge cases
   - Load tests to verify performance
   - Chaos engineering to test failure scenarios

## Extending with Custom Implementations

You can implement your own rate limiter by implementing the `RateLimiter` interface:

```go
type CustomRateLimiter struct {
    // Your implementation
}

func (c *CustomRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
    // Your rate limiting logic
    return true, nil
}

func (c *CustomRateLimiter) Reset(ctx context.Context, key string) error {
    // Your reset logic
    return nil
}
```

This allows integration with:
- External rate limiting services
- Custom algorithms
- Specialized storage backends
- Business-specific logic
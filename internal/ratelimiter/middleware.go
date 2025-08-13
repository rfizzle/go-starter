package ratelimiter

import (
	"go.uber.org/zap"
	"net/http"
)

// RateLimitMiddleware creates middleware that rate limits authentication attempts
func RateLimitMiddleware(rateLimiter RateLimiter, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			key := "auth_rate_limit:" + clientIP

			allowed, err := rateLimiter.Allow(r.Context(), key)
			if err != nil {
				logger.Error("rate limiter error", zap.String("client_ip", clientIP), zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				logger.Warn("rate limit exceeded", zap.String("client_ip", clientIP))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

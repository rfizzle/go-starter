package v1

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rfizzle/go-starter/internal/auth"
)

// AuditMiddleware returns a middleware that logs authentication events
func (s *service) AuditMiddleware(auditor auth.AuditLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a response writer that captures status code
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Serve the request
			next.ServeHTTP(rw, r)

			// Determine if this was an authentication attempt
			if isAuthRelatedRequest(r) {
				// Get user information if available
				userID, _ := s.subjectFromContext(r.Context())
				
				// Determine success based on status code
				success := rw.statusCode < 400

				// Log the event
				LogAuthEvent(
					r.Context(),
					auditor,
					"authentication_attempt",
					userID,
					GetClientIP(r),
					r.UserAgent(),
					success,
					nil,
					nil,
				)
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// isAuthRelatedRequest determines if a request is authentication-related
func isAuthRelatedRequest(r *http.Request) bool {
	// This is a simple heuristic - you might want to customize this
	// based on your specific authentication endpoints
	authPaths := []string{"/auth", "/login", "/logout", "/token"}
	
	for _, path := range authPaths {
		if strings.Contains(r.URL.Path, path) {
			return true
		}
	}
	
	// Also check if Authorization header is present
	return r.Header.Get("Authorization") != ""
}

// GetClientIP extracts the real client IP from the request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (most common)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// LogAuthEvent logs an authentication event using the provided auditor
func LogAuthEvent(ctx context.Context, auditor auth.AuditLogger, eventType, userID, clientIP, userAgent string, success bool, err error, metadata map[string]string) {
	event := auth.AuditEvent{
		EventType: eventType,
		UserID:    userID,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Success:   success,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  metadata,
	}

	if err != nil {
		event.Error = err.Error()
	}

	if logErr := auditor.LogEvent(ctx, event); logErr != nil {
		// Log to application logger as fallback
		// Note: This would need the logger to be passed in or made global
		_ = logErr // Suppress unused variable warning for now
	}
}

// Legacy audit function for backward compatibility
func (a *legacyAuth) AuditMiddleware(auditor auth.AuditLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a response writer that captures status code
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Serve the request
			next.ServeHTTP(rw, r)

			// Determine if this was an authentication attempt
			if isAuthRelatedRequest(r) {
				// Get user information if available
				userID, _ := subjectFromContext(r.Context(), a.contextKey)
				
				// Determine success based on status code
				success := rw.statusCode < 400

				// Log the event
				LogAuthEvent(
					r.Context(),
					auditor,
					"authentication_attempt",
					userID,
					GetClientIP(r),
					r.UserAgent(),
					success,
					nil,
					nil,
				)
			}
		})
	}
}
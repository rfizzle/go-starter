package auth

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// AuditEvent represents an authentication-related event
type AuditEvent struct {
	EventType string            `json:"event_type"`
	UserID    string            `json:"user_id,omitempty"`
	ClientIP  string            `json:"client_ip"`
	UserAgent string            `json:"user_agent"`
	Timestamp time.Time         `json:"timestamp"`
	Success   bool              `json:"success"`
	Error     string            `json:"error,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// AuditLogger interface for logging security events
type AuditLogger interface {
	LogEvent(ctx context.Context, event AuditEvent) error
}

// AuditMiddleware creates middleware that logs authentication events
func (a *Auth) AuditMiddleware(auditor AuditLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Create a response writer that captures status
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Try to get subject from context
			subject, err := a.subjectFromContext(r.Context())

			// Call next handler
			next.ServeHTTP(rw, r)

			// Log the event
			event := AuditEvent{
				EventType: "http_request",
				UserID:    subject,
				ClientIP:  getClientIP(r),
				UserAgent: r.UserAgent(),
				Timestamp: startTime,
				Success:   rw.status < 400,
				Metadata: map[string]string{
					"method":      r.Method,
					"path":        r.URL.Path,
					"status_code": http.StatusText(rw.status),
					"duration_ms": time.Since(startTime).String(),
				},
			}

			if err != nil {
				event.Error = err.Error()
			}

			if logErr := auditor.LogEvent(r.Context(), event); logErr != nil {
				a.logger.Error("failed to log audit event", zap.Error(logErr))
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
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

// LogAuthEvent logs authentication-specific events
func LogAuthEvent(ctx context.Context, auditor AuditLogger, eventType, userID, clientIP, userAgent string, success bool, err error, metadata map[string]string) {
	event := AuditEvent{
		EventType: eventType,
		UserID:    userID,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		Success:   success,
		Metadata:  metadata,
	}

	if err != nil {
		event.Error = err.Error()
	}

	if logErr := auditor.LogEvent(ctx, event); logErr != nil {
		// Log to application logger as fallback
		// Note: This would need the logger to be passed in or made global
	}
}

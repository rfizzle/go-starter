package middleware

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/go-chi/chi/v5/middleware"
)

func Logger(logger *zap.Logger, loggingPaths []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only log on defined paths (static requests can be logged via forward web server like nginx).
			// With no patterns configured there is no filter, so every request is logged. Otherwise
			// log when the request matches ANY configured pattern and skip only when none match.
			if len(loggingPaths) > 0 {
				var matched bool
				for _, path := range loggingPaths {
					if stringGlob(path, r.URL.Path) {
						matched = true
						break
					}
				}
				if !matched {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Start a timer for request duration
			start := time.Now()

			// Get the request ID
			var requestID string
			if reqID := r.Context().Value(middleware.RequestIDKey); reqID != nil {
				requestID = reqID.(string)
			}

			// Wrap the response writer
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			reqDuration := time.Since(start)
			fields := []zapcore.Field{
				zap.Int("status", ww.Status()),
				zap.Duration("took", reqDuration),
				zap.String("remote", r.RemoteAddr),
				zap.String("request", r.RequestURI),
				zap.String("method", r.Method),
				zap.Int("resp-bytes", ww.BytesWritten()),
			}
			// ContentLength is -1 when the request body size is unknown (e.g.
			// chunked transfer encoding); only log it when the client declared one.
			if r.ContentLength >= 0 {
				fields = append(fields, zap.Int64("req-bytes", r.ContentLength))
			}
			if requestID != "" {
				fields = append(fields, zap.String("request-id", requestID))
			}
			logger.Debug("request completed", fields...)
		})
	}
}

// stringGlob will test a string pattern, potentially containing globs, against a
// subject string. The result is a simple true/false, determining whether
// the glob pattern matched the subject text or not.
func stringGlob(pattern, subj string) bool {
	// Empty pattern can only match empty subject
	if pattern == "" {
		return subj == pattern
	}

	// If the pattern _is_ a glob, it matches everything
	if pattern == "*" {
		return true
	}

	parts := strings.Split(pattern, "*")

	if len(parts) == 1 {
		// No globs in pattern, so test for equality
		return subj == pattern
	}

	leadingGlob := strings.HasPrefix(pattern, "*")
	trailingGlob := strings.HasSuffix(pattern, "*")
	end := len(parts) - 1

	// Go over the leading parts and ensure they match.
	for i := 0; i < end; i++ {
		idx := strings.Index(subj, parts[i])

		switch i {
		case 0:
			// Check the first section. Requires special handling.
			if !leadingGlob && idx != 0 {
				return false
			}
		default:
			// Check that the middle parts match.
			if idx < 0 {
				return false
			}
		}

		// Trim evaluated text from subj as we loop over the pattern.
		subj = subj[idx+len(parts[i]):]
	}

	// Reached the last section. Requires special handling.
	return trailingGlob || strings.HasSuffix(subj, parts[end])
}

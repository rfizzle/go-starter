package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newBufferLogger returns a Debug-level logger that writes JSON entries to buf,
// so tests can assert whether a request was logged.
func newBufferLogger(buf *bytes.Buffer) *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&syncBuffer{Buffer: buf}), zapcore.DebugLevel)
	return zap.New(core)
}

// syncBuffer is a goroutine-safe zapcore.WriteSyncer backed by a bytes.Buffer.
type syncBuffer struct {
	*bytes.Buffer
	mu sync.Mutex
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Buffer.Write(p)
}

func (s *syncBuffer) Sync() error { return nil }

func TestLogger(t *testing.T) {
	tests := []struct {
		name         string
		loggingPaths []string
		requestPath  string
		wantLogged   bool
	}{
		{
			name:         "single pattern matches",
			loggingPaths: []string{"/api/*"},
			requestPath:  "/api/v1/users",
			wantLogged:   true,
		},
		{
			name:         "single pattern does not match",
			loggingPaths: []string{"/api/*"},
			requestPath:  "/static/app.js",
			wantLogged:   false,
		},
		{
			// Regression: with multiple patterns, a request matching the 2nd
			// (but not the 1st) pattern must still be logged. The old
			// implementation skipped on the first non-match.
			name:         "matches second of multiple patterns",
			loggingPaths: []string{"/api/*", "/admin/*"},
			requestPath:  "/admin/dashboard",
			wantLogged:   true,
		},
		{
			name:         "matches first of multiple patterns",
			loggingPaths: []string{"/api/*", "/admin/*"},
			requestPath:  "/api/v1/users",
			wantLogged:   true,
		},
		{
			name:         "matches none of multiple patterns",
			loggingPaths: []string{"/api/*", "/admin/*"},
			requestPath:  "/static/app.js",
			wantLogged:   false,
		},
		{
			// No configured patterns means no filter, so every request is logged.
			name:         "empty patterns logs everything",
			loggingPaths: []string{},
			requestPath:  "/static/app.js",
			wantLogged:   true,
		},
		{
			name:         "nil patterns logs everything",
			loggingPaths: nil,
			requestPath:  "/static/app.js",
			wantLogged:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := newBufferLogger(&buf)

			var called bool
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			handler := Logger(logger, tt.loggingPaths)(next)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			handler.ServeHTTP(rec, req)

			if !called {
				t.Fatal("next handler was not called")
			}

			logged := buf.Len() > 0
			if logged != tt.wantLogged {
				t.Errorf("logged = %v, want %v (log output: %q)", logged, tt.wantLogged, buf.String())
			}
		})
	}
}

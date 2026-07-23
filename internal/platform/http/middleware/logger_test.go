package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// TestLoggerEnrichesSizeFields verifies the request log records the response
// size (BytesWritten) and, when the client declares one, the request size
// (ContentLength).
func TestLoggerEnrichesSizeFields(t *testing.T) {
	const respBody = "hello world"

	tests := []struct {
		name        string
		body        string
		declareZero bool
		wantReqSize bool
		wantReqSz   int64
	}{
		{
			name:        "request with body logs req-bytes",
			body:        "payload-1234",
			wantReqSize: true,
			wantReqSz:   int64(len("payload-1234")),
		},
		{
			name:        "request without body omits req-bytes when length unknown",
			body:        "",
			wantReqSize: false,
		},
		{
			// A declared zero-length body is meaningful (ContentLength == 0),
			// so req-bytes must still be logged — this pins the `>= 0` boundary
			// against a regression to `> 0`.
			name:        "request with declared zero-length body logs req-bytes",
			body:        "",
			declareZero: true,
			wantReqSize: true,
			wantReqSz:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := newBufferLogger(&buf)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(respBody))
			})

			handler := Logger(logger, nil)(next)

			rec := httptest.NewRecorder()
			var req *http.Request
			switch {
			case tt.declareZero:
				// A POST with an explicitly declared empty body: ContentLength == 0.
				req = httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(""))
				req.ContentLength = 0
			case tt.body == "":
				// A GET with no body leaves ContentLength at -1 (unknown).
				req = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
				req.ContentLength = -1
			default:
				req = httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tt.body))
			}
			handler.ServeHTTP(rec, req)

			var entry map[string]any
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("failed to parse log entry %q: %v", buf.String(), err)
			}

			respBytes, ok := entry["resp-bytes"]
			if !ok {
				t.Fatalf("resp-bytes field missing from log entry: %q", buf.String())
			}
			if got := int(respBytes.(float64)); got != len(respBody) {
				t.Errorf("resp-bytes = %d, want %d", got, len(respBody))
			}

			reqBytes, hasReqBytes := entry["req-bytes"]
			if hasReqBytes != tt.wantReqSize {
				t.Fatalf("req-bytes present = %v, want %v (log output: %q)", hasReqBytes, tt.wantReqSize, buf.String())
			}
			if tt.wantReqSize {
				if got := int64(reqBytes.(float64)); got != tt.wantReqSz {
					t.Errorf("req-bytes = %d, want %d", got, tt.wantReqSz)
				}
			}
		})
	}
}

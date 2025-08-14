package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Mock audit logger for testing
type mockAuditLogger struct {
	events []AuditEvent
	err    error
}

func (m *mockAuditLogger) LogEvent(ctx context.Context, event AuditEvent) error {
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, event)
	return nil
}

// Mock rate limiter for testing
type mockRateLimiter struct {
	allowed map[string]bool
	err     error
}

func (m *mockRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	allowed, exists := m.allowed[key]
	if !exists {
		return true, nil // Default to allowing if not specified
	}
	return allowed, nil
}

func (m *mockRateLimiter) Reset(ctx context.Context, key string) error {
	delete(m.allowed, key)
	return nil
}

func TestAuditMiddleware(t *testing.T) {
	validator := &mockValidator{}
	logger := zap.NewNop()
	auth, err := NewAuth(validator, WithLogger(logger))
	assert.NoError(t, err)

	auditor := &mockAuditLogger{}
	middleware := auth.AuditMiddleware(auditor)

	t.Run("logs successful request", func(t *testing.T) {
		ctx, err := auth.newSubjectContext(context.Background(), "user123")
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/users", nil).WithContext(ctx)
		req.Header.Set("User-Agent", "TestAgent/1.0")
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Len(t, auditor.events, 1)
		event := auditor.events[0]
		assert.Equal(t, "http_request", event.EventType)
		assert.Equal(t, "user123", event.UserID)
		assert.Equal(t, "TestAgent/1.0", event.UserAgent)
		assert.True(t, event.Success)
		assert.Empty(t, event.Error)
		assert.Equal(t, "GET", event.Metadata["method"])
		assert.Equal(t, "/api/users", event.Metadata["path"])
	})

	t.Run("logs failed request", func(t *testing.T) {
		auditor.events = []AuditEvent{} // Reset

		req := httptest.NewRequest("POST", "/api/users", nil)
		req.Header.Set("User-Agent", "TestAgent/1.0")
		rr := httptest.NewRecorder()

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		})

		handler := middleware(nextHandler)
		handler.ServeHTTP(rr, req)

		assert.Len(t, auditor.events, 1)
		event := auditor.events[0]
		assert.Equal(t, "http_request", event.EventType)
		assert.Empty(t, event.UserID) // No authenticated user
		assert.False(t, event.Success)
		assert.NotEmpty(t, event.Error) // Should have error from missing subject
	})
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name             string
		setupRequest     func() *http.Request
		expectedIP       string
		expectedContains bool
	}{
		{
			name: "X-Forwarded-For header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("X-Forwarded-For", "203.0.113.1")
				req.RemoteAddr = "192.0.2.1:12345"
				return req
			},
			expectedIP: "203.0.113.1",
		},
		{
			name: "X-Real-IP header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("X-Real-IP", "203.0.113.2")
				req.RemoteAddr = "192.0.2.1:12345"
				return req
			},
			expectedIP: "203.0.113.2",
		},
		{
			name: "RemoteAddr fallback",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.RemoteAddr = "192.0.2.1:12345"
				return req
			},
			expectedIP: "192.0.2.1:12345",
		},
		{
			name: "X-Forwarded-For takes precedence",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("X-Forwarded-For", "203.0.113.1")
				req.Header.Set("X-Real-IP", "203.0.113.2")
				req.RemoteAddr = "192.0.2.1:12345"
				return req
			},
			expectedIP: "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			ip := getClientIP(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestLogAuthEvent(t *testing.T) {
	auditor := &mockAuditLogger{}

	LogAuthEvent(
		context.Background(),
		auditor,
		"login_success",
		"user123",
		"192.0.2.1",
		"TestAgent/1.0",
		true,
		nil,
		map[string]string{"method": "password"},
	)

	assert.Len(t, auditor.events, 1)
	event := auditor.events[0]
	assert.Equal(t, "login_success", event.EventType)
	assert.Equal(t, "user123", event.UserID)
	assert.Equal(t, "192.0.2.1", event.ClientIP)
	assert.Equal(t, "TestAgent/1.0", event.UserAgent)
	assert.True(t, event.Success)
	assert.Empty(t, event.Error)
	assert.Equal(t, "password", event.Metadata["method"])
}

func TestResponseWriter(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rr, status: http.StatusOK}

	// Test default status
	assert.Equal(t, http.StatusOK, rw.status)

	// Test WriteHeader
	rw.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, rw.status)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Test Write
	_, _ = rw.Write([]byte("test response"))
	assert.Equal(t, "test response", rr.Body.String())
}

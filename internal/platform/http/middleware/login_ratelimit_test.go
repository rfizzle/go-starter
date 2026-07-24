package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type recordingLimiter struct {
	allowed bool
	err     error
	keys    []string
}

func (limiter *recordingLimiter) Allow(_ context.Context, key string) (bool, error) {
	limiter.keys = append(limiter.keys, key)
	return limiter.allowed, limiter.err
}

func TestLoginRateLimit_UsesTrustedClientIdentity(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		forwardedFor   []string
		trustedProxies []string
		wantKey        string
	}{
		{
			name:           "direct peer ignores spoofed forwarding header",
			remoteAddr:     "192.0.2.10:1234",
			forwardedFor:   []string{"198.51.100.20"},
			trustedProxies: []string{"10.0.0.0/8"},
			wantKey:        "192.0.2.10",
		},
		{
			name:           "trusted single proxy supplies client",
			remoteAddr:     "10.0.0.10:1234",
			forwardedFor:   []string{"203.0.113.7"},
			trustedProxies: []string{"10.0.0.0/8"},
			wantKey:        "203.0.113.7",
		},
		{
			name:           "trusted multi hop chain selects first untrusted address from right",
			remoteAddr:     "10.0.0.20:1234",
			forwardedFor:   []string{"198.51.100.99, 203.0.113.7, 10.0.0.10"},
			trustedProxies: []string{"10.0.0.0/8"},
			wantKey:        "203.0.113.7",
		},
		{
			name:           "trusted proxy reads all forwarding header field lines",
			remoteAddr:     "10.0.0.20:1234",
			forwardedFor:   []string{"198.51.100.99", "10.0.0.10"},
			trustedProxies: []string{"10.0.0.0/8"},
			wantKey:        "198.51.100.99",
		},
		{
			name:           "malformed trusted chain falls back to immediate peer",
			remoteAddr:     "10.0.0.20:1234",
			forwardedFor:   []string{"not-an-ip"},
			trustedProxies: []string{"10.0.0.0/8"},
			wantKey:        "10.0.0.20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := &recordingLimiter{allowed: true}
			rateLimit, err := NewLoginRateLimit(limiter, tt.trustedProxies, time.Minute, zap.NewNop())
			require.NoError(t, err)

			nextCalled := false
			handler := rateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			}))
			request := httptest.NewRequest(http.MethodPost, loginPath, nil)
			request.RemoteAddr = tt.remoteAddr
			for _, value := range tt.forwardedFor {
				request.Header.Add("X-Forwarded-For", value)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			assert.Equal(t, http.StatusNoContent, response.Code)
			assert.True(t, nextCalled)
			assert.Equal(t, []string{tt.wantKey}, limiter.keys)
		})
	}
}

func TestLoginRateLimit_AppliesOnlyToLogin(t *testing.T) {
	limiter := &recordingLimiter{allowed: false}
	rateLimit, err := NewLoginRateLimit(limiter, nil, time.Minute, zap.NewNop())
	require.NoError(t, err)
	handler := rateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, limiter.keys)
}

func TestLoginRateLimit_DeniedResponseIncludesRetryMetadata(t *testing.T) {
	limiter := &recordingLimiter{allowed: false}
	rateLimit, err := NewLoginRateLimit(limiter, nil, 1500*time.Millisecond, zap.NewNop())
	require.NoError(t, err)
	handler := rateLimit(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("denied login reached next handler")
	}))

	request := httptest.NewRequest(http.MethodPost, loginPath, nil)
	request.RemoteAddr = "192.0.2.10:1234"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	assert.Equal(t, http.StatusTooManyRequests, response.Code)
	assert.Equal(t, "2", response.Header().Get("Retry-After"))
	assert.Equal(t, "application/json", response.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"code":429,"message":"Too Many Requests","status":"fail"}`, response.Body.String())
}

func TestLoginRateLimit_FailsClosedOnLimiterError(t *testing.T) {
	limiter := &recordingLimiter{allowed: true, err: errors.New("limiter unavailable")}
	rateLimit, err := NewLoginRateLimit(limiter, nil, time.Minute, zap.NewNop())
	require.NoError(t, err)
	handler := rateLimit(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("limiter error reached next handler")
	}))

	request := httptest.NewRequest(http.MethodPost, loginPath, nil)
	request.RemoteAddr = "192.0.2.10:1234"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Empty(t, response.Header().Get("Retry-After"))
}

func TestNewLoginRateLimit_ValidatesDependencies(t *testing.T) {
	limiter := &recordingLimiter{allowed: true}

	tests := []struct {
		name    string
		limiter requestLimiter
		window  time.Duration
		proxies []string
	}{
		{name: "missing limiter", window: time.Minute},
		{name: "nonpositive window", limiter: limiter},
		{name: "invalid proxy", limiter: limiter, window: time.Minute, proxies: []string{"not-a-network"}},
		{name: "noncanonical proxy", limiter: limiter, window: time.Minute, proxies: []string{"192.0.2.1/24"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rateLimit, err := NewLoginRateLimit(tt.limiter, tt.proxies, tt.window, zap.NewNop())
			assert.Error(t, err)
			assert.Nil(t, rateLimit)
		})
	}
}

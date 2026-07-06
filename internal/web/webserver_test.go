package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-chi/chi/v5"
)

// freePort asks the OS for an available TCP port on the loopback interface and
// returns it. Binding tests to an ephemeral port keeps them from colliding with
// the production default (8080) or with each other. The listener is closed
// before the port is reused, leaving a brief reuse window; acceptable here since
// these tests don't run in parallel and the loopback ephemeral range is large.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to reserve a free port: %v", err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

func TestNewWebserver(t *testing.T) {
	handler := http.NewServeMux()
	ws, err := New(handler)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if ws == nil {
		t.Fatal("Expected webserver instance, got nil")
	}
}

func TestWebserverStartAndStop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	handler := chi.NewRouter()
	handler.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	host := "127.0.0.1"
	port := freePort(t)

	ws, err := New(handler, WithHost(host), WithPort(port))
	assert.NoError(t, err)

	// Run drives both startup and graceful shutdown off the context: cancelling
	// ctx triggers the shutdown path, and runErr carries any run error back.
	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		runErr <- ws.Run(ctx)
	}()

	// Ensure the server is stopped even if an assertion below fails, then verify
	// a context-driven shutdown returns nil.
	defer func() {
		cancel()
		if err := <-runErr; err != nil {
			t.Errorf("Failed to stop webserver: %v", err)
		}
	}()

	url := fmt.Sprintf("http://%s:%d/api/v1/test", host, port)

	// Poll until the server accepts connections rather than sleeping a fixed
	// duration, so the test is neither flaky nor unnecessarily slow.
	var resp *http.Response
	deadline := time.Now().Add(5 * time.Second)
	for {
		resp, err = http.Get(url)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("Server did not become ready in time: %v", err)
		}
		time.Sleep(25 * time.Millisecond)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
}

func TestPprofGating(t *testing.T) {
	// pprof is a security-sensitive surface, so lock in that it stays off the
	// mux unless explicitly enabled and appears only when WithPprof(true) is set.
	tests := []struct {
		name       string
		enabled    bool
		wantStatus int
	}{
		{name: "disabled by default", enabled: false, wantStatus: http.StatusNotFound},
		{name: "enabled via option", enabled: true, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws, err := New(http.NewServeMux(), WithPprof(tt.enabled))
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
			rec := httptest.NewRecorder()
			ws.server.Handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestCORSOrigins(t *testing.T) {
	// The default is permissive (any origin echoed back); WithCORSOrigins should
	// restrict allowed origins so a disallowed origin gets no
	// Access-Control-Allow-Origin header.
	tests := []struct {
		name        string
		options     []func(*Webserver)
		origin      string
		wantAllowed bool
	}{
		{name: "default allows any origin", origin: "https://evil.example.com", wantAllowed: true},
		{
			name:        "restricted allows configured origin",
			options:     []func(*Webserver){WithCORSOrigins([]string{"https://app.example.com"})},
			origin:      "https://app.example.com",
			wantAllowed: true,
		},
		{
			name:        "restricted blocks other origin",
			options:     []func(*Webserver){WithCORSOrigins([]string{"https://app.example.com"})},
			origin:      "https://evil.example.com",
			wantAllowed: false,
		},
		{
			name:        "empty slice keeps permissive default",
			options:     []func(*Webserver){WithCORSOrigins(nil)},
			origin:      "https://evil.example.com",
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws, err := New(http.NewServeMux(), tt.options...)
			assert.NoError(t, err)

			// A CORS preflight is the reliable way to observe the middleware's
			// allow/deny decision via the response headers.
			req := httptest.NewRequest(http.MethodOptions, "/api/v1/test", nil)
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("Access-Control-Request-Method", http.MethodGet)
			rec := httptest.NewRecorder()
			ws.server.Handler.ServeHTTP(rec, req)

			allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
			if tt.wantAllowed {
				assert.NotEmpty(t, allowOrigin, "expected origin to be allowed")
			} else {
				assert.Empty(t, allowOrigin, "expected origin to be blocked")
			}
		})
	}
}

func TestReadHeaderTimeout(t *testing.T) {
	// A non-zero ReadHeaderTimeout on the http.Server is what closes the
	// Slowloris surface, so verify both the safe default and the override.
	t.Run("safe default when unset", func(t *testing.T) {
		ws, err := New(http.NewServeMux())
		assert.NoError(t, err)
		assert.Equal(t, 5*time.Second, ws.server.ReadHeaderTimeout)
	})

	t.Run("override via option", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithReadHeaderTimeout(2*time.Second))
		assert.NoError(t, err)
		assert.Equal(t, 2*time.Second, ws.server.ReadHeaderTimeout)
	})

	t.Run("non-positive keeps default", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithReadHeaderTimeout(0))
		assert.NoError(t, err)
		assert.Equal(t, 5*time.Second, ws.server.ReadHeaderTimeout)
	})
}

func TestMaxHeaderBytes(t *testing.T) {
	// An explicit MaxHeaderBytes on the http.Server bounds the memory a single
	// request's headers can force the server to buffer, so verify both the safe
	// default and the override.
	t.Run("safe default when unset", func(t *testing.T) {
		ws, err := New(http.NewServeMux())
		assert.NoError(t, err)
		assert.Equal(t, 1<<20, ws.server.MaxHeaderBytes)
	})

	t.Run("override via option", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithMaxHeaderBytes(2<<20))
		assert.NoError(t, err)
		assert.Equal(t, 2<<20, ws.server.MaxHeaderBytes)
	})

	t.Run("non-positive keeps default", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithMaxHeaderBytes(0))
		assert.NoError(t, err)
		assert.Equal(t, 1<<20, ws.server.MaxHeaderBytes)
	})
}

func TestWithHostAndPortOptions(t *testing.T) {
	handler := http.NewServeMux()
	host := "127.0.0.1"
	port := 9090
	ws, err := New(handler, WithHost(host), WithPort(port))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if ws == nil {
		t.Fatal("Expected webserver instance, got nil")
	}
	if ws.host != host {
		t.Errorf("Expected host %s, got %s", host, ws.host)
	}
	if ws.port != port {
		t.Errorf("Expected port %d, got %d", port, ws.port)
	}
}

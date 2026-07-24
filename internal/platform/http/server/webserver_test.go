package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-chi/chi/v5"
)

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

	ws, err := New(handler, WithHost("127.0.0.1"), WithPort(0))
	assert.NoError(t, err)

	// Run drives both startup and graceful shutdown off the context: cancelling
	// ctx triggers the shutdown path, and runErr carries any run error back.
	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		runErr <- ws.Run(ctx)
	}()

	stopped := false
	defer func() {
		if stopped {
			return
		}
		cancel()
		if err := <-runErr; err != nil {
			t.Errorf("Failed to stop webserver: %v", err)
		}
	}()

	var address net.Addr
	select {
	case address = <-ws.Started():
	case err := <-runErr:
		stopped = true
		t.Fatalf("Webserver.Run() returned before startup: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("Webserver.Run() did not report startup")
	}

	url := fmt.Sprintf("http://%s/api/v1/test", address)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("Failed to create readiness request: %v", err)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	cancel()
	if err := <-runErr; err != nil {
		t.Fatalf("Failed to stop webserver: %v", err)
	}
	stopped = true
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

func TestRequestTimeout(t *testing.T) {
	// The chi Timeout middleware bounds total request processing; verify the safe
	// default, an override, and that a non-positive value disables it.
	t.Run("safe default when unset", func(t *testing.T) {
		ws, err := New(http.NewServeMux())
		assert.NoError(t, err)
		assert.Equal(t, 60*time.Second, ws.requestTimeout)
	})

	t.Run("override via option", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithRequestTimeout(5*time.Second))
		assert.NoError(t, err)
		assert.Equal(t, 5*time.Second, ws.requestTimeout)
	})

	t.Run("non-positive disables the limit", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithRequestTimeout(0))
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(0), ws.requestTimeout)
	})

	t.Run("deadline cancels a slow handler", func(t *testing.T) {
		// A handler that outlives the request timeout should see its context
		// cancelled, which is what frees the connection under the middleware.
		var ctxErr error
		handler := chi.NewRouter()
		handler.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
				ctxErr = r.Context().Err()
			case <-time.After(2 * time.Second):
			}
			w.WriteHeader(http.StatusOK)
		})

		ws, err := New(handler, WithRequestTimeout(50*time.Millisecond))
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/slow", nil)
		rec := httptest.NewRecorder()
		ws.server.Handler.ServeHTTP(rec, req)

		assert.ErrorIs(t, ctxErr, context.DeadlineExceeded)
	})
}

func TestShutdownTimeout(t *testing.T) {
	// The shutdown timeout bounds how long graceful shutdown waits for in-flight
	// requests to drain, so verify both the safe default and the override.
	t.Run("safe default when unset", func(t *testing.T) {
		ws, err := New(http.NewServeMux())
		assert.NoError(t, err)
		assert.Equal(t, 15*time.Second, ws.shutdownTimeout)
	})

	t.Run("override via option", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithShutdownTimeout(30*time.Second))
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Second, ws.shutdownTimeout)
	})

	t.Run("non-positive keeps default", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithShutdownTimeout(0))
		assert.NoError(t, err)
		assert.Equal(t, 15*time.Second, ws.shutdownTimeout)
	})
}

func TestTLSConfiguration(t *testing.T) {
	// TLS is off by default (proxy-terminated posture) and only activates when
	// both a cert and key path are supplied. When it does, a safe TLS 1.2 floor
	// is installed unless overridden.
	t.Run("disabled by default", func(t *testing.T) {
		ws, err := New(http.NewServeMux())
		assert.NoError(t, err)
		assert.False(t, ws.tlsEnabled())
		assert.Nil(t, ws.server.TLSConfig)
	})

	t.Run("enabled installs safe default floor", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithTLS("cert.pem", "key.pem"))
		assert.NoError(t, err)
		assert.True(t, ws.tlsEnabled())
		if assert.NotNil(t, ws.server.TLSConfig) {
			assert.Equal(t, uint16(tls.VersionTLS12), ws.server.TLSConfig.MinVersion)
		}
	})

	t.Run("missing key path is a no-op", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithTLS("cert.pem", ""))
		assert.NoError(t, err)
		assert.False(t, ws.tlsEnabled())
		assert.Nil(t, ws.server.TLSConfig)
	})

	t.Run("custom config overrides default", func(t *testing.T) {
		custom := &tls.Config{MinVersion: tls.VersionTLS13}
		ws, err := New(http.NewServeMux(), WithTLS("cert.pem", "key.pem"), WithTLSConfig(custom))
		assert.NoError(t, err)
		if assert.NotNil(t, ws.server.TLSConfig) {
			assert.Equal(t, uint16(tls.VersionTLS13), ws.server.TLSConfig.MinVersion)
		}
	})

	t.Run("nil custom config keeps default", func(t *testing.T) {
		ws, err := New(http.NewServeMux(), WithTLS("cert.pem", "key.pem"), WithTLSConfig(nil))
		assert.NoError(t, err)
		if assert.NotNil(t, ws.server.TLSConfig) {
			assert.Equal(t, uint16(tls.VersionTLS12), ws.server.TLSConfig.MinVersion)
		}
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

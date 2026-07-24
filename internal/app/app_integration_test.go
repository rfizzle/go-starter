//go:build integration

package app

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rfizzle/go-starter/internal/platform/config"
	"github.com/rfizzle/go-starter/internal/platform/database/postgrestest"
)

const lifecycleWait = 10 * time.Second

type applicationFixture struct {
	application *App
	baseURL     string
	client      *http.Client
	cancel      context.CancelFunc
	runErr      chan error
	stopped     bool
}

func TestApplicationLifecycle(t *testing.T) {
	harness := postgrestest.New(t)

	t.Run("serves probes and rejects unauthenticated requests before graceful shutdown", func(t *testing.T) {
		fixture := newApplicationFixture(t, harness.Database(t).URI)

		fixture.requireStatus(t, "/api/v1/healthz/liveness", http.StatusOK)
		fixture.requireStatus(t, "/api/v1/healthz/readiness", http.StatusOK)
		fixture.requireStatus(t, "/api/v1/me", http.StatusUnauthorized)

		fixture.stop(t)
		if fixture.application.Ready() {
			t.Fatal("App.Ready() = true after graceful shutdown")
		}

		fixture.client.CloseIdleConnections()
		request, err := http.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			fixture.baseURL+"/api/v1/healthz/liveness",
			nil,
		)
		if err != nil {
			t.Fatalf("create post-shutdown request: %v", err)
		}
		if response, err := fixture.client.Do(request); err == nil {
			_ = response.Body.Close()
			t.Fatal("liveness request succeeded after graceful shutdown")
		}
	})

	t.Run("listener startup failure is returned without reporting readiness", func(t *testing.T) {
		blocker, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("bind occupied listener: %v", err)
		}
		defer func() { _ = blocker.Close() }()

		address, ok := blocker.Addr().(*net.TCPAddr)
		if !ok {
			t.Fatalf("listener address type = %T, want *net.TCPAddr", blocker.Addr())
		}

		cfg := lifecycleConfig(harness.Database(t).URI)
		cfg.Webserver.Address = "127.0.0.1"
		cfg.Webserver.Port = address.Port

		application, err := New(cfg)
		if err != nil {
			t.Fatalf("New() error: %v", err)
		}

		runErr := make(chan error, 1)
		go func() {
			runErr <- application.Run(t.Context())
		}()

		select {
		case err := <-runErr:
			if err == nil {
				t.Fatal("App.Run() error = nil, want listener startup failure")
			}
			if !strings.Contains(err.Error(), "issue opening webserver listener") {
				t.Fatalf("App.Run() error = %q, want listener startup context", err)
			}
		case address := <-application.Started():
			t.Fatalf("App.Started() reported %s after listener startup failure", address)
		case <-time.After(lifecycleWait):
			t.Fatal("App.Run() did not return listener startup failure")
		}

		if application.Ready() {
			t.Fatal("App.Ready() = true after listener startup failure")
		}
	})
}

func newApplicationFixture(t *testing.T, databaseURI string) *applicationFixture {
	t.Helper()

	application, err := New(lifecycleConfig(databaseURI))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	runCtx, cancel := context.WithCancel(t.Context())
	fixture := &applicationFixture{
		application: application,
		client: &http.Client{
			Timeout: lifecycleWait,
		},
		cancel: cancel,
		runErr: make(chan error, 1),
	}
	go func() {
		fixture.runErr <- application.Run(runCtx)
	}()

	select {
	case address := <-application.Started():
		fixture.baseURL = "http://" + address.String()
	case err := <-fixture.runErr:
		fixture.stopped = true
		cancel()
		t.Fatalf("App.Run() returned before startup: %v", err)
	case <-time.After(lifecycleWait):
		cancel()
		t.Fatal("App.Run() did not report startup")
	}

	t.Cleanup(func() {
		if !fixture.stopped {
			fixture.stop(t)
		}
	})
	return fixture
}

func (fixture *applicationFixture) requireStatus(t *testing.T, path string, want int) {
	t.Helper()

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		fixture.baseURL+path,
		nil,
	)
	if err != nil {
		t.Fatalf("create GET %s: %v", path, err)
	}
	response, err := fixture.client.Do(request)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer func() { _ = response.Body.Close() }()
	if _, err := io.Copy(io.Discard, response.Body); err != nil {
		t.Fatalf("read GET %s response: %v", path, err)
	}
	if response.StatusCode != want {
		t.Fatalf("GET %s status = %d, want %d", path, response.StatusCode, want)
	}
}

func (fixture *applicationFixture) stop(t *testing.T) {
	t.Helper()
	if fixture.stopped {
		return
	}

	fixture.cancel()
	select {
	case err := <-fixture.runErr:
		fixture.stopped = true
		if err != nil {
			t.Fatalf("App.Run() graceful shutdown error: %v", err)
		}
	case <-time.After(lifecycleWait):
		t.Fatal("App.Run() did not complete graceful shutdown")
	}
}

func lifecycleConfig(databaseURI string) *config.Application {
	return &config.Application{
		Environment: "production",
		Webserver: config.Webserver{
			Address:         "127.0.0.1",
			Port:            0,
			RequestTimeout:  5,
			ShutdownTimeout: 5,
		},
		Database: config.Database{
			URI:          databaseURI,
			PingTimeout:  5,
			MaxConns:     4,
			CloseTimeout: 5,
		},
		Auth: config.Auth{
			JWTSecret: strings.Repeat("lifecycle-secret-", 2),
			JWTTTL:    3600,
		},
		Logging: config.Logging{
			Level: "info",
		},
	}
}

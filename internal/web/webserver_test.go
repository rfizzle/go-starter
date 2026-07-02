package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
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

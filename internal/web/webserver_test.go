package web

import (
	"net/http"
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

	ws, err := New(handler)
	assert.NoError(t, err)
	go func() {
		if err := ws.Start(); err != nil {
			t.Errorf("Failed to start webserver: %v", err)
		}
	}()

	// Allow some time for the server to start
	time.Sleep(5 * time.Second)

	resp, err := http.Get("http://localhost:8080/api/v1/test")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	if err := ws.Stop(); err != nil {
		t.Fatalf("Failed to stop webserver: %v", err)
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

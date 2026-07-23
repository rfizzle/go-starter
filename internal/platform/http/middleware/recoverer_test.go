package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	dto "github.com/prometheus/client_model/go"
)

// counterValue reads the current value of a Prometheus counter for assertions.
func counterValue(t *testing.T) float64 {
	t.Helper()
	var m dto.Metric
	if err := panicsTotal.Write(&m); err != nil {
		t.Fatalf("reading panicsTotal: %v", err)
	}
	return m.GetCounter().GetValue()
}

func TestRecoverer(t *testing.T) {
	t.Run("recovers panic, logs, increments metric, and returns 500", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newBufferLogger(&buf)

		next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("boom")
		})

		handler := Recoverer(logger)(next)

		before := counterValue(t)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

		// The middleware must not let the panic escape.
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}

		if got := counterValue(t) - before; got != 1 {
			t.Errorf("panicsTotal delta = %v, want 1", got)
		}

		out := buf.String()
		if out == "" {
			t.Fatal("expected a log entry, got none")
		}
		if !strings.Contains(out, "recovered from panic") {
			t.Errorf("log missing panic message: %q", out)
		}
		if !strings.Contains(out, "boom") {
			t.Errorf("log missing panic value: %q", out)
		}
		if !strings.Contains(out, "stack") {
			t.Errorf("log missing stack trace field: %q", out)
		}
		// With no RequestID middleware in the chain, the field must be omitted.
		if strings.Contains(out, "request-id") {
			t.Errorf("log should not contain request-id when none is set: %q", out)
		}
	})

	t.Run("preserves an already-written status", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newBufferLogger(&buf)

		// Handler commits a 200 and writes a partial body before panicking.
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("partial"))
			panic("boom")
		})

		handler := Recoverer(logger)(next)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		handler.ServeHTTP(rec, req)

		// The middleware must not overwrite the committed 200 with a 500.
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d (already-written status must be preserved)", rec.Code, http.StatusOK)
		}
		if !strings.Contains(buf.String(), "recovered from panic") {
			t.Errorf("panic should still be logged: %q", buf.String())
		}
	})

	t.Run("logs request id when present", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newBufferLogger(&buf)

		next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("boom")
		})

		// Chain RequestID before Recoverer so a request ID is in context.
		handler := middleware.RequestID(Recoverer(logger)(next))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		handler.ServeHTTP(rec, req)

		if !strings.Contains(buf.String(), "request-id") {
			t.Errorf("log missing request-id field: %q", buf.String())
		}
	})

	t.Run("passes through when no panic", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newBufferLogger(&buf)

		var called bool
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		handler := Recoverer(logger)(next)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		handler.ServeHTTP(rec, req)

		if !called {
			t.Fatal("next handler was not called")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if buf.Len() != 0 {
			t.Errorf("expected no log output, got %q", buf.String())
		}
	})

	t.Run("re-panics on http.ErrAbortHandler", func(t *testing.T) {
		defer func() {
			if rvr := recover(); rvr != http.ErrAbortHandler {
				t.Errorf("recover() = %v, want http.ErrAbortHandler", rvr)
			}
		}()

		next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic(http.ErrAbortHandler)
		})

		handler := Recoverer(nil)(next)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		handler.ServeHTTP(rec, req)

		t.Fatal("expected http.ErrAbortHandler to propagate")
	})
}

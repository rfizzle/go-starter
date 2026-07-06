package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMaxBytes(t *testing.T) {
	// readAllHandler drains the request body and returns 200 on success. When the
	// body exceeds the MaxBytesReader limit, io.ReadAll fails and it writes 413.
	readAllHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	t.Run("allows a body within the limit", func(t *testing.T) {
		handler := MaxBytes(16)(readAllHandler)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader("small"))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("rejects a body over the limit", func(t *testing.T) {
		handler := MaxBytes(4)(readAllHandler)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader("this is too large"))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
		}
	})

	t.Run("surfaces a MaxBytesError past the limit", func(t *testing.T) {
		// The wrapped body must fail with *http.MaxBytesError specifically, so
		// downstream handlers can distinguish an oversized body from other read
		// errors and map it to the right status.
		var readErr error
		next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			_, readErr = io.ReadAll(r.Body)
		})
		handler := MaxBytes(4)(next)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader("this is too large"))
		handler.ServeHTTP(rec, req)

		var maxErr *http.MaxBytesError
		if !errors.As(readErr, &maxErr) {
			t.Errorf("read error = %v, want *http.MaxBytesError", readErr)
		}
	})

	t.Run("allows a body exactly at the limit", func(t *testing.T) {
		// MaxBytesReader permits up to n bytes and errors only on the (n+1)th, so
		// a body whose length equals the limit must read cleanly.
		const limit = 5
		handler := MaxBytes(limit)(readAllHandler)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(strings.Repeat("x", limit)))
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d (a body at the limit must be allowed)", rec.Code, http.StatusOK)
		}
	})

	t.Run("disables the limit when n is non-positive", func(t *testing.T) {
		// A non-positive limit (0 or negative) must return the next handler
		// untouched, so even a large body reads cleanly.
		for _, n := range []int64{0, -1} {
			handler := MaxBytes(n)(readAllHandler)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(strings.Repeat("x", 1<<20)))
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("n=%d: status = %d, want %d (non-positive limit should not restrict the body)", n, rec.Code, http.StatusOK)
			}
		}
	})
}

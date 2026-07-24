package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFileServerPreservesApplicationRoutes(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	static := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	handler := FileServer(static)(next)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{name: "api root", path: "/api", wantStatus: http.StatusNoContent},
		{name: "api route", path: "/api/v1/users", wantStatus: http.StatusNoContent},
		{name: "metrics", path: "/metrics", wantStatus: http.StatusNoContent},
		{name: "metrics subpath", path: "/metrics/missing", wantStatus: http.StatusNoContent},
		{name: "pprof route", path: "/debug/pprof/", wantStatus: http.StatusNoContent},
		{name: "debug route", path: "/debug/vars", wantStatus: http.StatusNoContent},
		{name: "frontend root", path: "/", wantStatus: http.StatusAccepted},
		{name: "frontend asset", path: "/assets/app.js", wantStatus: http.StatusAccepted},
		{name: "api lookalike", path: "/api-docs", wantStatus: http.StatusAccepted},
		{name: "metrics lookalike", path: "/metrics-dashboard", wantStatus: http.StatusAccepted},
		{name: "debug lookalike", path: "/debugger", wantStatus: http.StatusAccepted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("GET %s status = %d, want %d", tt.path, response.Code, tt.wantStatus)
			}
		})
	}
}

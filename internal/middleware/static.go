package middleware

import (
	"net/http"
	"strings"
)

func FileServer(static http.Handler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasPrefix(r.URL.Path, "/api"):
				next.ServeHTTP(w, r)
			case strings.HasPrefix(r.URL.Path, "/metrics"):
				next.ServeHTTP(w, r)
			case strings.HasPrefix(r.URL.Path, "/debug/"):
				next.ServeHTTP(w, r)
			default:
				static.ServeHTTP(w, r)
			}
		})
	}
}

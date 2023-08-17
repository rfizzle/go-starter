package middleware

import (
	"net/http"
	"strings"
)

func fileServer(static http.Handler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				next.ServeHTTP(w, r)
			} else if strings.HasPrefix(r.URL.Path, "/healthz") {
				next.ServeHTTP(w, r)
			} else {
				static.ServeHTTP(w, r)
			}
		})
	}
}

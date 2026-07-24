package middleware

import (
	"net/http"
	"strings"
)

// FileServer routes requests outside the application's reserved HTTP
// namespaces to the embedded static handler. Keeping the namespace check here
// ensures a frontend fallback cannot shadow application endpoints.
func FileServer(static http.Handler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case isPathWithin(r.URL.Path, "/api"),
				isPathWithin(r.URL.Path, "/metrics"),
				isPathWithin(r.URL.Path, "/debug"):
				next.ServeHTTP(w, r)
			default:
				static.ServeHTTP(w, r)
			}
		})
	}
}

func isPathWithin(requestPath, namespace string) bool {
	return requestPath == namespace || strings.HasPrefix(requestPath, namespace+"/")
}

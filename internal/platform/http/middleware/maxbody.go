package middleware

import "net/http"

// MaxBytes returns middleware that caps the size of a request body at n bytes.
// It wraps each request body in an http.MaxBytesReader (via
// http.MaxBytesHandler), so reading past the limit returns a *http.MaxBytesError
// from the body Read. This bounds the memory an oversized POST/PUT can force the
// server to buffer. The status the client sees depends on the downstream
// handler: a handler that surfaces the error directly responds 413 Request
// Entity Too Large, while ogen's default error handler maps the decode failure
// to 400 Bad Request.
//
// A non-positive n disables the limit and passes requests through untouched, so
// callers can opt out explicitly.
func MaxBytes(n int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if n <= 0 {
			return next
		}
		return http.MaxBytesHandler(next, n)
	}
}

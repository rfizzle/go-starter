// Package middleware provides the HTTP middleware the application's server
// mounts: structured request logging, panic recovery, request-body size
// limits, reserved-path-aware static file serving, and trusted-proxy-aware
// login rate limiting. Every constructor returns a Constructor, so the results
// mount directly on a chi router or compose by hand.
//
// It is technical platform infrastructure and carries no business meaning: it
// must not import internal/domain or internal/modules.
package middleware

import "net/http"

// Constructor wraps an http.Handler with one piece of middleware. It is the
// shape chi's Use accepts, so a constructor built here can be mounted directly
// on a router or composed by hand.
type Constructor func(http.Handler) http.Handler

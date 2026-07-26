package middleware

import "net/http"

// Constructor wraps an http.Handler with one piece of middleware. It is the
// shape chi's Use accepts, so a constructor built here can be mounted directly
// on a router or composed by hand.
type Constructor func(http.Handler) http.Handler

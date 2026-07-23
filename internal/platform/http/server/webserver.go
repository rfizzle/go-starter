// Package server implements the application's HTTP server: middleware stack,
// route mounting, and graceful lifecycle management.
//
// # TLS
//
// By default this server speaks plain HTTP and assumes TLS is terminated
// upstream by a reverse proxy, load balancer, or ingress controller (the
// common posture for containerised/Kubernetes deployments). It does not
// redirect HTTP→HTTPS or set HSTS; that belongs to the terminating layer.
//
// For standalone deployments where the Go process is itself internet-facing,
// enable in-process TLS termination with WithTLS (certificate + key paths).
// That negotiates a TLS 1.2 protocol floor by default; WithTLSConfig overrides
// the full *tls.Config for advanced setups (mutual TLS, a raised MinVersion, a
// pinned cipher-suite list). When TLS is enabled the server calls
// ListenAndServeTLS; otherwise it calls ListenAndServe.
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/pprof"
	"sync/atomic"
	"time"

	"github.com/rfizzle/go-starter/internal/platform/http/middleware"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-faster/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Webserver struct {
	host              string
	port              int
	server            *http.Server
	logger            *zap.Logger
	fileserver        http.Handler
	readHeaderTimeout time.Duration
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	requestTimeout    time.Duration
	shutdownTimeout   time.Duration
	maxHeaderBytes    int
	maxBodyBytes      int64
	pprofEnabled      bool
	corsOrigins       []string
	// TLS termination. When tlsCertFile and tlsKeyFile are both set (or a
	// tlsConfig is supplied) the server terminates TLS in-process via
	// ListenAndServeTLS; otherwise it serves plain HTTP and assumes an upstream
	// proxy terminates TLS. tlsConfig, when nil, is populated with a safe
	// default (TLS 1.2 floor) whenever cert/key paths are provided.
	tlsCertFile string
	tlsKeyFile  string
	tlsConfig   *tls.Config
}

func New(apiHandlers http.Handler, options ...func(c *Webserver)) (*Webserver, error) {
	isReady := &atomic.Bool{}
	isReady.Store(false)

	// Setup a new app with default values
	newApp := &Webserver{
		host: "0.0.0.0",
		port: 8080,
		// A short header-read deadline closes the Slowloris attack surface (a
		// client trickling headers to hold connections open) by default, even
		// when the operator leaves the other timeouts unset.
		readHeaderTimeout: 5 * time.Second,
		// Bound total request-processing time so a slow handler (or a client that
		// stalls mid-request) can't hold a connection indefinitely. Applied via
		// chi's Timeout middleware, which cancels the request context at the
		// deadline. Set to 0 via WithRequestTimeout to disable.
		requestTimeout: 60 * time.Second,
		// Bound how long a graceful shutdown waits for in-flight requests to drain
		// before the server is forced closed, so a wedged handler can't hang the
		// process indefinitely on SIGTERM.
		shutdownTimeout: 15 * time.Second,
		// Cap total request header size at 1 MiB (Go's own default) so an
		// oversized-header request can't force the server to buffer unbounded
		// memory. Set explicitly rather than relying on the implicit default so
		// the limit is visible and tunable.
		maxHeaderBytes: 1 << 20,
		// Cap each request body at 10 MiB so an oversized POST/PUT can't force
		// the server to buffer unbounded memory. Downstream handlers that need to
		// accept larger payloads can raise this via WithMaxBodyBytes.
		maxBodyBytes: 10 << 20,
		// Permissive default: allows any origin. This is a placeholder — restrict
		// to your front-end origins via WithCORSOrigins before production.
		corsOrigins: []string{"https://*", "http://*"},
		logger:      zap.NewNop(),
	}

	// Apply options
	for _, option := range options {
		option(newApp)
	}

	// Setup middleware
	mux := chi.NewMux()
	// A non-positive request timeout disables the middleware entirely; otherwise
	// chi cancels the request context once the deadline elapses.
	if newApp.requestTimeout > 0 {
		mux.Use(chimiddleware.Timeout(newApp.requestTimeout))
	}
	mux.Use(chimiddleware.RequestID)
	mux.Use(chimiddleware.RealIP)
	mux.Use(middleware.Recoverer(newApp.logger))
	mux.Use(middleware.MaxBytes(newApp.maxBodyBytes))
	mux.Use(middleware.Logger(newApp.logger, []string{"/api/*"}))
	mux.Use(middleware.FileServer(newApp.fileserver))
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   newApp.corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value isn't ignored by any of the major browsers
	}))
	mux.Handle("/metrics", promhttp.Handler())

	// pprof exposes process internals (goroutine stacks, heap, CPU profiles) and
	// is an information-leak and DoS surface, so it stays off the public mux
	// unless explicitly enabled. Gate behind config rather than mounting it
	// unconditionally.
	if newApp.pprofEnabled {
		mux.Route("/debug/pprof", func(mux chi.Router) {
			mux.HandleFunc("/", pprof.Index)
			mux.HandleFunc("/cmdline", pprof.Cmdline)
			mux.HandleFunc("/profile", pprof.Profile)
			mux.HandleFunc("/symbol", pprof.Symbol)
			mux.HandleFunc("/trace", pprof.Trace)
		})
	}

	// Mount the API handlers
	mux.Mount("/api/v1", http.StripPrefix("/api/v1", apiHandlers))

	// When in-process TLS is requested via cert/key paths but no explicit
	// tls.Config was supplied, install a safe default that refuses protocols
	// below TLS 1.2. Go's own default MinVersion is already TLS 1.2, but we set
	// it explicitly so the floor is visible and survives future stdlib changes.
	if newApp.tlsConfig == nil && newApp.tlsCertFile != "" && newApp.tlsKeyFile != "" {
		newApp.tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	// Create a new HTTP server with the handler
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", newApp.host, newApp.port),
		ReadHeaderTimeout: newApp.readHeaderTimeout,
		WriteTimeout:      newApp.writeTimeout,
		ReadTimeout:       newApp.readTimeout,
		IdleTimeout:       newApp.idleTimeout,
		MaxHeaderBytes:    newApp.maxHeaderBytes,
		TLSConfig:         newApp.tlsConfig,
		ErrorLog:          zap.NewStdLog(newApp.logger),
		Handler:           mux,
	}

	// Set the server on the app
	newApp.server = srv

	return newApp, nil
}

// tlsEnabled reports whether the server should terminate TLS in-process. It is
// true when both a certificate and key path were configured via WithTLS.
func (application *Webserver) tlsEnabled() bool {
	return application.tlsCertFile != "" && application.tlsKeyFile != ""
}

// Run starts the HTTP server and blocks until ctx is cancelled or the server
// fails. On cancellation it performs a graceful shutdown, bounded by the
// configured shutdown timeout. It returns nil on a clean shutdown so it composes
// with an errgroup without treating an intentional stop as an error.
func (application *Webserver) Run(ctx context.Context) error {
	// ListenAndServe blocks, so run it in a goroutine and surface any startup
	// error through a buffered channel (buffered so the goroutine never leaks
	// if we return via ctx.Done() first).
	serveErr := make(chan error, 1)
	go func() {
		// ListenAndServeTLS with empty cert/key paths reads the certificate from
		// the server's TLSConfig; we only take the TLS path when explicit paths
		// were configured, so pass them through directly.
		var err error
		if application.tlsEnabled() {
			err = application.server.ListenAndServeTLS(application.tlsCertFile, application.tlsKeyFile)
		} else {
			err = application.server.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- errors.Wrap(err, "issue serving webserver")
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		// Use a fresh context: the one we were handed is already cancelled, so
		// deriving the shutdown deadline from it would abort immediately.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), application.shutdownTimeout)
		defer cancel()

		if err := application.server.Shutdown(shutdownCtx); err != nil {
			return errors.Wrap(err, "issue stopping webserver")
		}

		return nil
	}
}

func WithHost(host string) func(*Webserver) {
	return func(w *Webserver) {
		w.host = host
	}
}

func WithPort(port int) func(*Webserver) {
	return func(w *Webserver) {
		w.port = port
	}
}

func WithLogger(logger *zap.Logger) func(*Webserver) {
	return func(w *Webserver) {
		w.logger = logger
	}
}

// WithReadHeaderTimeout bounds how long the server waits for a client to send
// its request headers, mitigating Slowloris-style attacks. A non-positive
// timeout is ignored so the safe default remains in place.
func WithReadHeaderTimeout(timeout time.Duration) func(*Webserver) {
	return func(w *Webserver) {
		if timeout > 0 {
			w.readHeaderTimeout = timeout
		}
	}
}

func WithReadTimeout(timeout time.Duration) func(*Webserver) {
	return func(w *Webserver) {
		w.readTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) func(*Webserver) {
	return func(w *Webserver) {
		w.writeTimeout = timeout
	}
}

func WithIdleTimeout(timeout time.Duration) func(*Webserver) {
	return func(w *Webserver) {
		w.idleTimeout = timeout
	}
}

// WithRequestTimeout bounds total request-processing time via chi's Timeout
// middleware, cancelling the request context once the deadline elapses so a slow
// handler or stalled client can't hold a connection indefinitely. A non-positive
// timeout disables the middleware entirely (unbounded processing), overriding
// the safe 60s default rather than preserving it.
func WithRequestTimeout(timeout time.Duration) func(*Webserver) {
	return func(w *Webserver) {
		w.requestTimeout = timeout
	}
}

// WithShutdownTimeout bounds how long a graceful shutdown waits for in-flight
// requests to drain before the server is forced closed, so a wedged handler
// can't hang the process on SIGTERM. A non-positive timeout is ignored so the
// safe 15s default remains in place.
func WithShutdownTimeout(timeout time.Duration) func(*Webserver) {
	return func(w *Webserver) {
		if timeout > 0 {
			w.shutdownTimeout = timeout
		}
	}
}

// WithMaxHeaderBytes caps the total size of request headers the server will
// read, bounding the memory a single request can force the server to buffer. A
// non-positive value is ignored so the safe 1 MiB default remains in place.
func WithMaxHeaderBytes(maxBytes int) func(*Webserver) {
	return func(w *Webserver) {
		if maxBytes > 0 {
			w.maxHeaderBytes = maxBytes
		}
	}
}

// WithMaxBodyBytes caps the size of each request body the server will read,
// bounding the memory a single request can force the server to buffer. Unlike
// WithMaxHeaderBytes, a non-positive value disables the limit entirely
// (unbounded bodies), overriding the safe 10 MiB default rather than preserving
// it — 0 means "no limit", not "reject everything".
func WithMaxBodyBytes(maxBytes int64) func(*Webserver) {
	return func(w *Webserver) {
		w.maxBodyBytes = maxBytes
	}
}

func WithFileServer(fs http.Handler) func(*Webserver) {
	return func(w *Webserver) {
		w.fileserver = fs
	}
}

// WithPprof toggles the /debug/pprof/* profiling endpoints. They are disabled by
// default because they expose process internals and are unauthenticated on the
// public mux; enable only in trusted environments or behind separate access
// controls.
func WithPprof(enabled bool) func(*Webserver) {
	return func(w *Webserver) {
		w.pprofEnabled = enabled
	}
}

// WithTLS enables in-process TLS termination using the PEM-encoded certificate
// and key at the given paths. Use this only for standalone deployments where
// the Go process is internet-facing; when TLS is terminated upstream by a
// reverse proxy or load balancer (the default posture) leave it unset. Both
// paths must be non-empty for TLS to activate — if either is empty the option
// is a no-op and the server continues to serve plain HTTP. Unless a full
// tls.Config is supplied via WithTLSConfig, the server negotiates a TLS 1.2
// protocol floor.
func WithTLS(certFile, keyFile string) func(*Webserver) {
	return func(w *Webserver) {
		if certFile != "" && keyFile != "" {
			w.tlsCertFile = certFile
			w.tlsKeyFile = keyFile
		}
	}
}

// WithTLSConfig sets the tls.Config used when TLS is enabled via WithTLS,
// overriding the default (a TLS 1.2 floor). Use it for advanced setups such as
// mutual TLS, a raised MinVersion, or a pinned cipher-suite list. A nil config
// is ignored so the safe default remains in place. This option only takes
// effect when WithTLS has also supplied certificate and key paths.
func WithTLSConfig(cfg *tls.Config) func(*Webserver) {
	return func(w *Webserver) {
		if cfg != nil {
			w.tlsConfig = cfg
		}
	}
}

// WithCORSOrigins sets the origins permitted by the CORS middleware. Each entry
// may contain a "*" wildcard (e.g. "https://*.example.com"). An empty slice is
// ignored so the permissive default remains in place. Restrict this to your
// front-end origins before deploying to production.
func WithCORSOrigins(origins []string) func(*Webserver) {
	return func(w *Webserver) {
		if len(origins) > 0 {
			w.corsOrigins = origins
		}
	}
}

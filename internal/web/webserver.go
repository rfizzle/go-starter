package web

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"sync/atomic"
	"time"

	"github.com/rfizzle/go-starter/internal/middleware"

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
	maxHeaderBytes    int
	pprofEnabled      bool
	corsOrigins       []string
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
		// Cap total request header size at 1 MiB (Go's own default) so an
		// oversized-header request can't force the server to buffer unbounded
		// memory. Set explicitly rather than relying on the implicit default so
		// the limit is visible and tunable.
		maxHeaderBytes: 1 << 20,
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
	mux.Use(chimiddleware.Timeout(60 * time.Second))
	mux.Use(chimiddleware.RequestID)
	mux.Use(chimiddleware.RealIP)
	mux.Use(chimiddleware.Recoverer)
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

	// Create a new HTTP server with the handler
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", newApp.host, newApp.port),
		ReadHeaderTimeout: newApp.readHeaderTimeout,
		WriteTimeout:      newApp.writeTimeout,
		ReadTimeout:       newApp.readTimeout,
		IdleTimeout:       newApp.idleTimeout,
		MaxHeaderBytes:    newApp.maxHeaderBytes,
		ErrorLog:          zap.NewStdLog(newApp.logger),
		Handler:           mux,
	}

	// Set the server on the app
	newApp.server = srv

	return newApp, nil
}

// shutdownTimeout bounds how long a graceful shutdown waits for in-flight
// requests to drain before the server is forced closed.
const shutdownTimeout = 15 * time.Second

// Run starts the HTTP server and blocks until ctx is cancelled or the server
// fails. On cancellation it performs a graceful shutdown, bounded by
// shutdownTimeout. It returns nil on a clean shutdown so it composes with an
// errgroup without treating an intentional stop as an error.
func (application *Webserver) Run(ctx context.Context) error {
	// ListenAndServe blocks, so run it in a goroutine and surface any startup
	// error through a buffered channel (buffered so the goroutine never leaks
	// if we return via ctx.Done() first).
	serveErr := make(chan error, 1)
	go func() {
		if err := application.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
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

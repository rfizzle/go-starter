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
	host         string
	port         int
	server       *http.Server
	logger       *zap.Logger
	fileserver   http.Handler
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
}

func New(apiHandlers http.Handler, options ...func(c *Webserver)) (*Webserver, error) {
	isReady := &atomic.Bool{}
	isReady.Store(false)

	// Setup a new app with default values
	newApp := &Webserver{
		host:   "0.0.0.0",
		port:   8080,
		logger: zap.NewNop(),
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
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value isn't ignored by any of the major browsers
	}))
	mux.Handle("/metrics", promhttp.Handler())
	mux.Route("/debug/pprof", func(mux chi.Router) {
		mux.HandleFunc("/", pprof.Index)
		mux.HandleFunc("/cmdline", pprof.Cmdline)
		mux.HandleFunc("/profile", pprof.Profile)
		mux.HandleFunc("/symbol", pprof.Symbol)
		mux.HandleFunc("/trace", pprof.Trace)
	})

	// Mount the API handlers
	mux.Mount("/api/v1", http.StripPrefix("/api/v1", apiHandlers))

	// Create a new HTTP server with the handler
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", newApp.host, newApp.port),
		WriteTimeout: time.Second * newApp.writeTimeout,
		ReadTimeout:  time.Second * newApp.readTimeout,
		IdleTimeout:  time.Second * newApp.idleTimeout,
		ErrorLog:     zap.NewStdLog(newApp.logger),
		Handler:      mux,
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

func WithReadTimeout(timeout int) func(*Webserver) {
	return func(w *Webserver) {
		w.readTimeout = time.Duration(timeout)
	}
}

func WithWriteTimeout(timeout int) func(*Webserver) {
	return func(w *Webserver) {
		w.writeTimeout = time.Duration(timeout)
	}
}

func WithIdleTimeout(timeout int) func(*Webserver) {
	return func(w *Webserver) {
		w.idleTimeout = time.Duration(timeout)
	}
}

func WithFileServer(fs http.Handler) func(*Webserver) {
	return func(w *Webserver) {
		w.fileserver = fs
	}
}

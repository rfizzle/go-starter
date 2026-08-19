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
// enable in-process TLS termination by setting Config.TLSCertFile and
// Config.TLSKeyFile. The server then negotiates a TLS 1.2 protocol floor and
// serves with ServeTLS; otherwise it calls Serve.
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/rfizzle/go-starter/internal/platform/http/middleware"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-faster/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config holds the webserver settings. Every field has the same zero-value
// rule: an unset (zero, empty, or non-positive) value selects the default
// documented on the field, which applyDefaults installs before construction.
type Config struct {
	// Host is the interface to bind. Empty binds every interface.
	Host string
	// Port is the TCP port to listen on. 0 lets the operating system choose.
	Port int
	// Logger receives request, recovery, and http.Server logs. Nil discards
	// them.
	Logger *zap.Logger
	// FileServer serves requests outside the reserved application routes. Nil
	// serves no static files.
	FileServer http.Handler
	// ReadHeaderTimeout bounds how long a client may take to send its request
	// headers, closing the Slowloris surface. Default 5s.
	ReadHeaderTimeout time.Duration
	// ReadTimeout, WriteTimeout, and IdleTimeout are passed to http.Server; a
	// non-positive value keeps Go's own behaviour (no deadline, and idle
	// falling back to read).
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	// RequestTimeout bounds total request-processing time via chi's Timeout
	// middleware, which cancels the request context at the deadline so a slow
	// handler or stalled client cannot hold a connection indefinitely.
	// Default 60s.
	RequestTimeout time.Duration
	// ShutdownTimeout bounds how long a graceful shutdown waits for in-flight
	// requests to drain before the server is forced closed. Default 15s.
	ShutdownTimeout time.Duration
	// MaxHeaderBytes caps the total request header size the server reads,
	// bounding the memory one request can force it to buffer. Default 1 MiB.
	MaxHeaderBytes int
	// MaxBodyBytes caps each request body the server reads. Default 10 MiB.
	MaxBodyBytes int64
	// PprofEnabled mounts /debug/pprof/* on the public mux. Off by default:
	// the endpoints are unauthenticated and expose process internals.
	PprofEnabled bool
	// CORSOrigins are the origins permitted by the CORS middleware; entries may
	// contain a "*" wildcard. The default allows any origin and is a
	// placeholder to restrict before production.
	CORSOrigins []string
	// TLSCertFile and TLSKeyFile enable in-process TLS termination (with a TLS
	// 1.2 floor) when both are set. Leave them empty when a reverse proxy
	// terminates TLS — the default posture.
	TLSCertFile string
	TLSKeyFile  string
}

// applyDefaults fills every unset field with its documented default.
func (cfg *Config) applyDefaults() {
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
	if cfg.ReadHeaderTimeout <= 0 {
		cfg.ReadHeaderTimeout = 5 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 60 * time.Second
	}
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = 15 * time.Second
	}
	if cfg.MaxHeaderBytes <= 0 {
		cfg.MaxHeaderBytes = 1 << 20
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 10 << 20
	}
	if len(cfg.CORSOrigins) == 0 {
		cfg.CORSOrigins = []string{"https://*", "http://*"}
	}
}

// tlsEnabled reports whether the server should terminate TLS in-process. It is
// true only when both a certificate and a key path are configured.
func (cfg Config) tlsEnabled() bool {
	return cfg.TLSCertFile != "" && cfg.TLSKeyFile != ""
}

type Webserver struct {
	cfg     Config
	started chan net.Addr
	server  *http.Server
}

// New builds the middleware stack and http.Server from cfg after applying the
// documented defaults. The server is not started until Run is called.
func New(apiHandlers http.Handler, cfg Config) (*Webserver, error) {
	cfg.applyDefaults()

	mux := chi.NewMux()
	mux.Use(chimiddleware.Timeout(cfg.RequestTimeout))
	mux.Use(chimiddleware.RequestID)
	mux.Use(middleware.Recoverer(cfg.Logger))
	mux.Use(middleware.MaxBytes(cfg.MaxBodyBytes))
	mux.Use(middleware.Logger(cfg.Logger, []string{"/api/*"}))
	mux.Use(middleware.FileServer(cfg.FileServer))
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value isn't ignored by any of the major browsers
	}))
	mux.Handle("/metrics", promhttp.Handler())

	// pprof exposes process internals (goroutine stacks, heap, CPU profiles) and
	// is an information-leak and DoS surface, so it stays off the public mux
	// unless explicitly enabled.
	if cfg.PprofEnabled {
		mux.Route("/debug/pprof", func(mux chi.Router) {
			mux.HandleFunc("/", pprof.Index)
			mux.HandleFunc("/cmdline", pprof.Cmdline)
			mux.HandleFunc("/profile", pprof.Profile)
			mux.HandleFunc("/symbol", pprof.Symbol)
			mux.HandleFunc("/trace", pprof.Trace)
		})
	}

	mux.Mount("/api/v1", http.StripPrefix("/api/v1", apiHandlers))

	// Go's own default MinVersion is already TLS 1.2, but set it explicitly so
	// the floor is visible and survives future stdlib changes.
	var tlsConfig *tls.Config
	if cfg.tlsEnabled() {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	return &Webserver{
		cfg: cfg,
		// Buffered so Run can publish the bound address without depending on a
		// receiver being scheduled. A Webserver is a single-run lifecycle
		// component, so exactly one address is published.
		started: make(chan net.Addr, 1),
		server: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			IdleTimeout:       cfg.IdleTimeout,
			MaxHeaderBytes:    cfg.MaxHeaderBytes,
			TLSConfig:         tlsConfig,
			ErrorLog:          zap.NewStdLog(cfg.Logger),
			Handler:           mux,
		},
	}, nil
}

// Run starts the HTTP server and blocks until ctx is cancelled or the server
// fails. On cancellation it performs a graceful shutdown, bounded by the
// configured shutdown timeout. It returns nil on a clean shutdown so it composes
// with an errgroup without treating an intentional stop as an error.
func (application *Webserver) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", application.server.Addr)
	if err != nil {
		return errors.Wrap(err, "issue opening webserver listener")
	}
	defer func() { _ = listener.Close() }()

	// Serve blocks, so run it in a goroutine and surface any startup or runtime
	// error through a buffered channel. The listener wrapper publishes startup
	// only after http.Server has entered its accept loop; callers can therefore
	// synchronize on Started without polling or racing to reuse a "free" port.
	accepting := make(chan struct{})
	var acceptingOnce sync.Once
	listener = &acceptNotificationListener{
		Listener: listener,
		notify: func() {
			acceptingOnce.Do(func() {
				application.started <- listener.Addr()
				close(accepting)
			})
		},
	}

	serveErr := make(chan error, 1)
	go func() {
		var err error
		if application.cfg.tlsEnabled() {
			err = application.server.ServeTLS(listener, application.cfg.TLSCertFile, application.cfg.TLSKeyFile)
		} else {
			err = application.server.Serve(listener)
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- errors.Wrap(err, "issue serving webserver")
			return
		}
		serveErr <- nil
	}()

	select {
	case <-accepting:
		// The server owns and is actively accepting on the listener.
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		// Cancellation won before Serve entered its accept loop. Closing the
		// listener prevents a late Serve call from escaping shutdown.
		_ = listener.Close()
		<-serveErr
		return nil
	}

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		// Use a fresh context: the one we were handed is already cancelled, so
		// deriving the shutdown deadline from it would abort immediately.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), application.cfg.ShutdownTimeout)
		defer cancel()

		if err := application.server.Shutdown(shutdownCtx); err != nil {
			return errors.Wrap(err, "issue stopping webserver")
		}

		return <-serveErr
	}
}

// Started reports the address once the HTTP server has entered its accept
// loop. Webserver instances are single-use and publish exactly one address.
func (application *Webserver) Started() <-chan net.Addr {
	return application.started
}

type acceptNotificationListener struct {
	net.Listener
	notify func()
}

func (listener *acceptNotificationListener) Accept() (net.Conn, error) {
	listener.notify()
	return listener.Listener.Accept()
}

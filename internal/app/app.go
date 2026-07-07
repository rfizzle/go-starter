package app

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/auth"
	"github.com/rfizzle/go-starter/internal/controller"
	"github.com/rfizzle/go-starter/internal/observability"
	"github.com/rfizzle/go-starter/internal/storage/postgres"
	"github.com/rfizzle/go-starter/internal/user"
	"github.com/rfizzle/go-starter/internal/web"
	"github.com/rfizzle/go-starter/pkg/config"
	staticweb "github.com/rfizzle/go-starter/web"
)

type App struct {
	cfg    *config.Application
	ready  *atomic.Bool
	logger *zap.Logger
}

func New(cfg *config.Application) (*App, error) {
	isReady := &atomic.Bool{}
	isReady.Store(false)

	return &App{
		cfg:    cfg,
		ready:  isReady,
		logger: zap.NewNop(),
	}, nil
}

// Run starts the application and blocks until ctx is cancelled (e.g. an OS
// signal) or a long-lived component fails. All components derive from ctx, so
// the first failure cancels the group and drives a coordinated graceful
// shutdown.
func (application *App) Run(ctx context.Context) error {
	// Install OpenTelemetry providers before anything instrumented runs so the
	// generated ogen server's spans and metrics are exported from the first
	// request. When observability is disabled this is a no-op and the global
	// providers stay no-op. The shutdown flush runs on a fresh context because
	// ctx is already cancelled by the time we unwind (see the deferred call),
	// and it runs last (defer ordering) so late spans from draining requests are
	// still captured.
	otelShutdown, err := observability.Init(ctx, application.cfg)
	if err != nil {
		return fmt.Errorf("observability.Init(): %w", err)
	}
	defer func() {
		// Bound the flush so a wedged exporter connection can't hang shutdown.
		// Reuse the webserver's shutdown budget, falling back to a safe default
		// when it is left unset so we never derive a zero (already-expired)
		// deadline.
		flushTimeout := time.Duration(application.cfg.Webserver.ShutdownTimeout) * time.Second
		if flushTimeout <= 0 {
			flushTimeout = 15 * time.Second
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), flushTimeout)
		defer cancel()
		if err := otelShutdown(shutdownCtx); err != nil {
			application.logger.Warn("opentelemetry shutdown did not complete cleanly", zap.Error(err))
		}
	}()

	// Create the connection pool explicitly so the application owns it and can
	// share the exact same pool between the storage.Client (postgres.Client)
	// and the user domain repository. The pool is sized from config rather than
	// left at the pgx defaults (4 × NumCPU), which are frequently wrong under
	// load. pgxpool connects lazily, so the explicit Ping below is what makes an
	// unreachable database fail startup.
	pool, err := postgres.NewPool(ctx, application.cfg.Database.URI, postgres.PoolConfig{
		MaxConns:                 application.cfg.Database.MaxConns,
		MinConns:                 application.cfg.Database.MinConns,
		MaxConnLifetimeSeconds:   application.cfg.Database.MaxConnLifetime,
		MaxConnIdleTimeSeconds:   application.cfg.Database.MaxConnIdleTime,
		HealthCheckPeriodSeconds: application.cfg.Database.HealthCheckPeriod,
	})
	if err != nil {
		return fmt.Errorf("postgres.NewPool(): %w", err)
	}

	// The pool outlives the component group: it is closed only after every
	// component has returned, so in-flight requests can still reach the database
	// while the webserver drains. The close is bounded so a connection wedged
	// mid-query cannot hang shutdown indefinitely; on timeout the pool is
	// abandoned (its close still finishes in the background) and we log it.
	defer func() {
		closeTimeout := time.Duration(application.cfg.Database.CloseTimeout) * time.Second
		if err := postgres.ClosePool(pool, closeTimeout); err != nil {
			application.logger.Warn("database pool close did not complete cleanly", zap.Error(err))
		}
	}()

	// Expose pgxpool statistics (acquired/idle/total, empty-acquire count, …) on
	// the same default registry the /metrics endpoint serves, so connection-pool
	// exhaustion is observable before queries start timing out. The collector
	// reads pool.Stat() lazily on each scrape, so its lifetime is tied to the
	// pool: register here, unregister on shutdown so a subsequent Run (e.g. in a
	// test) does not collide with a stale collector on the process-wide registry.
	poolCollector := postgres.NewPoolCollector(pool)
	if err := prometheus.Register(poolCollector); err != nil {
		application.logger.Warn("could not register database pool metrics collector", zap.Error(err))
	} else {
		defer prometheus.Unregister(poolCollector)
	}

	store, err := postgres.Open(
		ctx,
		postgres.WithPool(pool),
		postgres.WithPingTimeout(application.cfg.Database.PingTimeout),
		postgres.WithLogger(application.logger),
	)
	if err != nil {
		return fmt.Errorf("postgres.Open(): %w", err)
	}

	if err := store.Ping(ctx); err != nil {
		return fmt.Errorf("store.Ping(): %w", err)
	}

	// Apply embedded, forward-only migrations before serving traffic so the
	// schema the user repository (and sqlc-generated queries) expect is present.
	if err := store.Migrate(postgres.Migrations); err != nil {
		return fmt.Errorf("store.Migrate(): %w", err)
	}

	// The user domain shares the same pool: sqlc-generated queries and squirrel
	// both execute against it, and WithinTx swaps in a pgx.Tx for atomic work.
	userRepo := user.NewRepository(pool)

	// Build the JWT validator that authenticates protected operations. When no
	// secret is configured the validator is nil, and the SecurityHandler fails
	// protected requests closed rather than accepting forged tokens.
	var validator auth.TokenValidator
	if application.cfg.Auth.JWTSecret != "" {
		jwtOpts := []auth.JWTOption{}
		if application.cfg.Auth.JWTIssuer != "" {
			jwtOpts = append(jwtOpts, auth.WithIssuer(application.cfg.Auth.JWTIssuer))
		}
		if application.cfg.Auth.JWTAudience != "" {
			jwtOpts = append(jwtOpts, auth.WithAudience(application.cfg.Auth.JWTAudience))
		}
		if application.cfg.Auth.JWTTTL > 0 {
			jwtOpts = append(jwtOpts, auth.WithTTL(time.Duration(application.cfg.Auth.JWTTTL)*time.Second))
		}

		jwtValidator, err := auth.NewJWT([]byte(application.cfg.Auth.JWTSecret), jwtOpts...)
		if err != nil {
			return fmt.Errorf("auth.NewJWT(): %w", err)
		}
		validator = jwtValidator
	} else {
		application.logger.Warn("no auth.jwt-secret configured; protected endpoints will reject all requests")
	}

	// Configure the controllers
	controllerOpts := controller.Options{
		ReadyState: application.ready,
		Store:      store,
		Users:      userRepo,
		Validator:  validator,
	}
	appController, err := controller.NewController(&controllerOpts)
	if err != nil {
		return fmt.Errorf("controller.NewController(): %w", err)
	}

	// Configure the API server. The controller doubles as the ogen
	// SecurityHandler, authenticating bearer tokens for protected operations.
	appServer, err := api.NewServer(appController, appController)
	if err != nil {
		return fmt.Errorf("api.NewServer(): %w", err)
	}

	// Configure static file server
	staticFs := staticweb.AssetHandler("/", "")

	// Configure application webserver
	webServer, err := web.New(
		appServer,
		web.WithHost(application.cfg.Webserver.Address),
		web.WithPort(application.cfg.Webserver.Port),
		web.WithReadHeaderTimeout(time.Duration(application.cfg.Webserver.ReadHeaderTimeout)*time.Second),
		web.WithReadTimeout(time.Duration(application.cfg.Webserver.ReadTimeout)*time.Second),
		web.WithWriteTimeout(time.Duration(application.cfg.Webserver.WriteTimeout)*time.Second),
		web.WithIdleTimeout(time.Duration(application.cfg.Webserver.IdleTimeout)*time.Second),
		web.WithRequestTimeout(time.Duration(application.cfg.Webserver.RequestTimeout)*time.Second),
		web.WithShutdownTimeout(time.Duration(application.cfg.Webserver.ShutdownTimeout)*time.Second),
		web.WithMaxHeaderBytes(application.cfg.Webserver.MaxHeaderBytes),
		web.WithMaxBodyBytes(application.cfg.Webserver.MaxBodyBytes),
		web.WithLogger(application.logger),
		web.WithFileServer(staticFs),
		web.WithPprof(application.cfg.Webserver.PprofEnabled),
		web.WithCORSOrigins(application.cfg.Webserver.CORSAllowedOrigins),
		// No-op unless both paths are set; otherwise TLS is assumed to be
		// terminated upstream by a reverse proxy.
		web.WithTLS(application.cfg.Webserver.TLSCertFile, application.cfg.Webserver.TLSKeyFile),
	)
	if err != nil {
		return fmt.Errorf("server.New(): %w", err)
	}

	// Run every long-lived component under a shared context. errgroup cancels
	// the derived context as soon as any component returns an error, so a
	// failing webserver (or, later, a worker) triggers shutdown of the rest
	// instead of leaving the process half-dead.
	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return webServer.Run(groupCtx)
	})

	// Set the ready flag to true
	application.ready.Store(true)

	// Log that we are serving the application
	application.logger.Info(fmt.Sprintf("application started on %s:%d", application.cfg.Webserver.Address, application.cfg.Webserver.Port))

	// Block until the group's context is cancelled (signal) or a component
	// fails. Mark the app not-ready as soon as we begin unwinding.
	err = group.Wait()
	application.ready.Store(false)
	application.logger.Info("application stopped")

	return err
}

func (application *App) Ready() bool {
	return application.ready.Load()
}

func (application *App) SetReady(ready bool) {
	application.ready.Store(ready)
}

func (application *App) Logger() *zap.Logger {
	return application.logger
}

func (application *App) SetLogger(logger *zap.Logger) {
	application.logger = logger
}

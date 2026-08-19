// Package app is the application's composition root. It constructs the
// concrete platform infrastructure (configuration, database, observability,
// HTTP server, rate limiting), adapts the business modules to the narrow
// consumer-owned interfaces the transport edge declares, and owns the
// resulting process lifecycle: startup ordering, readiness, and coordinated
// shutdown.
//
// It is the only internal package that may depend on modules, platform, and
// the generated transport edge at once. Nothing imports it except the serve
// command in cmd.
package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/errgroup"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/auth"
	"github.com/rfizzle/go-starter/internal/controller"
	"github.com/rfizzle/go-starter/internal/modules/credential"
	"github.com/rfizzle/go-starter/internal/modules/user"
	"github.com/rfizzle/go-starter/internal/platform/config"
	"github.com/rfizzle/go-starter/internal/platform/database/postgres"
	"github.com/rfizzle/go-starter/internal/platform/health"
	httpmiddleware "github.com/rfizzle/go-starter/internal/platform/http/middleware"
	httpserver "github.com/rfizzle/go-starter/internal/platform/http/server"
	"github.com/rfizzle/go-starter/internal/platform/observability"
	"github.com/rfizzle/go-starter/internal/platform/ratelimiter"
	staticweb "github.com/rfizzle/go-starter/web"
)

type App struct {
	cfg     *config.Application
	ready   *atomic.Bool
	logger  *zap.Logger
	started chan net.Addr
	// rateLimitClock is a deterministic test seam. Production leaves it nil so
	// the limiter uses time.Now.
	rateLimitClock func() time.Time
}

type appReadiness struct {
	lifecycle    *atomic.Bool
	dependencies controller.ReadinessChecker
}

var _ controller.ReadinessChecker = (*appReadiness)(nil)

func (readiness *appReadiness) Ready(ctx context.Context) bool {
	if readiness == nil || readiness.lifecycle == nil || !readiness.lifecycle.Load() {
		return false
	}
	return readiness.dependencies != nil && readiness.dependencies.Ready(ctx)
}

// credentialAuthenticator adapts internal/modules/credential.Service to
// controller.Authenticator, translating its ErrInvalidCredentials sentinel to
// the controller's own bounded sentinel so controller does not import a
// concrete business module (module-platform-boundaries.md, rule 6). It always
// returns the user's email as the subject, matching the convention
// GetCurrentUser uses to resolve a token subject back to a stored user.
type credentialAuthenticator struct {
	service *credential.Service
}

var _ controller.Authenticator = (*credentialAuthenticator)(nil)

func (a *credentialAuthenticator) Authenticate(ctx context.Context, email, password string) (string, error) {
	authenticated, err := a.service.Authenticate(ctx, email, password)
	if err != nil {
		if errors.Is(err, credential.ErrInvalidCredentials) {
			return "", controller.ErrInvalidCredentials
		}
		return "", err
	}
	return authenticated.Email, nil
}

func New(cfg *config.Application) (*App, error) {
	isReady := &atomic.Bool{}
	isReady.Store(false)

	application := &App{
		cfg:     cfg,
		ready:   isReady,
		logger:  zap.NewNop(),
		started: make(chan net.Addr, 1),
	}

	return application, nil
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
	// share the exact same pool between migrations and the user repository. The
	// pool is sized from config rather than left at the pgx defaults (4 ×
	// NumCPU), which are frequently wrong under load. pgxpool connects lazily,
	// so the explicit Ping below is what makes an unreachable database fail
	// startup.
	pool, err := postgres.NewPool(ctx, application.cfg.Database.URI, postgres.PoolConfigFromDatabase(application.cfg.Database))
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

	readinessTimeout := time.Duration(application.cfg.Database.PingTimeout) * time.Second
	if readinessTimeout <= 0 {
		readinessTimeout = postgres.DefaultPingTimeoutSeconds * time.Second
	}
	dependencyReadiness, err := health.New(
		readinessTimeout,
		health.Check{Name: "postgres", Run: pool.Ping},
	)
	if err != nil {
		return fmt.Errorf("health.New(): %w", err)
	}
	if err := dependencyReadiness.Check(ctx); err != nil {
		return fmt.Errorf("dependency readiness: %w", err)
	}

	// Startup migration is an explicit deployment policy. The safe default is
	// disabled so multiple replicas do not all perform schema work during a
	// rollout; deployments can run `migrate up` as a one-shot job instead.
	// When explicitly enabled, migration still completes before the HTTP server
	// is constructed, and the PostgreSQL advisory lock serializes replicas.
	migrate := func(ctx context.Context) error {
		return postgres.Migrate(ctx, pool, application.logger, postgres.Migrations)
	}
	if err := migrateOnStartup(ctx, application.cfg.Database.AutoMigrate, migrate); err != nil {
		return err
	}

	// The user domain shares the same pool: sqlc-generated queries and squirrel
	// both execute against it, and WithinTx swaps in a pgx.Tx for atomic work.
	userRepo := user.NewRepository(pool)

	// The credential service verifies login passwords against the user
	// repository's authentication-only projection; it has no JWT dependency of
	// its own. userRepo satisfies its package-private credentialReader port
	// structurally (CredentialByEmail) without this package importing that
	// interface.
	credentialService, err := credential.NewService(userRepo, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("credential.NewService(): %w", err)
	}

	// Build the JWT validator that authenticates protected operations. When no
	// secret is configured the validator is nil, and the SecurityHandler fails
	// protected requests closed rather than accepting forged tokens.
	var validator auth.TokenValidator
	var tokenGenerator auth.Generator
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
		tokenGenerator = jwtValidator
	} else {
		application.logger.Warn("no auth.jwt-secret configured; protected endpoints will reject all requests")
	}

	// Configure the controllers
	controllerOpts := controller.Options{
		Readiness: &appReadiness{
			lifecycle:    application.ready,
			dependencies: dependencyReadiness,
		},
		Users:         userRepo,
		Validator:     validator,
		Authenticator: &credentialAuthenticator{service: credentialService},
		Tokens:        tokenGenerator,
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

	rateLimitWindow := time.Duration(application.cfg.RateLimit.Window) * time.Second
	loginLimiter, err := ratelimiter.NewLimiter(ratelimiter.RateLimitConfig{
		MaxRequests:     application.cfg.RateLimit.MaxRequests,
		Window:          rateLimitWindow,
		MaxKeys:         application.cfg.RateLimit.MaxKeys,
		CleanupInterval: time.Duration(application.cfg.RateLimit.CleanupInterval) * time.Second,
		KeyPrefix:       "login:",
		Clock:           application.rateLimitClock,
	})
	if err != nil {
		return fmt.Errorf("ratelimiter.NewLimiter(): %w", err)
	}
	defer loginLimiter.Close()

	if err := prometheus.Register(loginLimiter); err != nil {
		application.logger.Warn("could not register login rate limiter metrics collector", zap.Error(err))
	} else {
		defer prometheus.Unregister(loginLimiter)
	}

	loginRateLimit, err := httpmiddleware.NewLoginRateLimit(
		loginLimiter,
		application.cfg.RateLimit.TrustedProxies,
		rateLimitWindow,
		application.logger,
	)
	if err != nil {
		return fmt.Errorf("middleware.NewLoginRateLimit(): %w", err)
	}
	apiHandler := loginRateLimit(appServer)

	// Configure static file server
	staticFs := staticweb.AssetHandler("/", "")

	// Configure application webserver
	webCfg := application.cfg.Webserver
	webServer, err := httpserver.New(apiHandler, httpserver.Config{
		Host:              webCfg.Address,
		Port:              webCfg.Port,
		Logger:            application.logger,
		FileServer:        staticFs,
		ReadHeaderTimeout: time.Duration(webCfg.ReadHeaderTimeout) * time.Second,
		ReadTimeout:       time.Duration(webCfg.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(webCfg.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(webCfg.IdleTimeout) * time.Second,
		RequestTimeout:    time.Duration(webCfg.RequestTimeout) * time.Second,
		ShutdownTimeout:   time.Duration(webCfg.ShutdownTimeout) * time.Second,
		MaxHeaderBytes:    webCfg.MaxHeaderBytes,
		MaxBodyBytes:      webCfg.MaxBodyBytes,
		PprofEnabled:      webCfg.PprofEnabled,
		CORSOrigins:       webCfg.CORSAllowedOrigins,
		// TLS activates only when both paths are set; otherwise TLS is assumed
		// to be terminated upstream by a reverse proxy.
		TLSCertFile: webCfg.TLSCertFile,
		TLSKeyFile:  webCfg.TLSKeyFile,
	})
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

	var address net.Addr
	select {
	case address = <-webServer.Started():
		// Publish readiness only after the listener is bound and the HTTP server
		// has entered its accept loop. The buffered signal lets fixtures use the
		// exact bound address without polling.
		application.setReady(true)
		application.started <- address
	case <-groupCtx.Done():
		err = group.Wait()
		application.setReady(false)
		return err
	}

	// Log that we are serving the application
	application.logger.Info("application started", zap.Stringer("address", address))

	// Block until the group's context is cancelled (signal) or a component
	// fails. Mark the app not-ready as soon as we begin unwinding.
	err = group.Wait()
	application.setReady(false)
	application.logger.Info("application stopped")

	return err
}

// migrateOnStartup runs migrate only when startup migration is enabled, so the
// policy decision stays testable without a database.
func migrateOnStartup(ctx context.Context, enabled bool, migrate func(context.Context) error) error {
	if !enabled {
		return nil
	}
	if err := migrate(ctx); err != nil {
		return fmt.Errorf("postgres.Migrate(): %w", err)
	}
	return nil
}

func (application *App) Ready() bool {
	return application.ready.Load()
}

// Started reports the bound listener address after dependency checks complete
// and the HTTP server has entered its accept loop. App instances are single-use
// and publish exactly one address.
func (application *App) Started() <-chan net.Addr {
	return application.started
}

func (application *App) SetReady(ready bool) {
	application.setReady(ready)
}

// setReady updates the readiness flag and logs genuine state transitions so
// readiness probe flaps are traceable. Swap reports the previous value
// atomically, so a write that does not change the flag (e.g. repeated
// not-ready during shutdown) is silent and only true→false / false→true edges
// are logged.
func (application *App) setReady(ready bool) {
	previous := application.ready.Swap(ready)
	if previous == ready {
		return
	}

	application.logger.Info("readiness state changed",
		zap.Bool("ready", ready),
		zap.Bool("previous", previous),
	)
}

func (application *App) Logger() *zap.Logger {
	return application.logger
}

func (application *App) SetLogger(logger *zap.Logger) {
	application.logger = logger
}

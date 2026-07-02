package app

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/controller"
	"github.com/rfizzle/go-starter/internal/storage/postgres"
	"github.com/rfizzle/go-starter/internal/web"
	"github.com/rfizzle/go-starter/pkg/config"
	staticweb "github.com/rfizzle/go-starter/web"
)

type App struct {
	cfg    *config.Application
	state  *state
	ready  *atomic.Bool
	logger *zap.Logger
}

func New(cfg *config.Application) (*App, error) {
	isReady := &atomic.Bool{}
	isReady.Store(false)

	return &App{
		cfg: cfg,
		state: &state{
			mutex:      sync.Mutex{},
			StartTime:  time.Now(),
			ErrorCount: 0,
		},
		ready:  isReady,
		logger: zap.NewNop(),
	}, nil
}

// Run starts the application and blocks until ctx is cancelled (e.g. an OS
// signal) or a long-lived component fails. All components derive from ctx, so
// the first failure cancels the group and drives a coordinated graceful
// shutdown.
func (application *App) Run(ctx context.Context) error {
	// Connect to the database and verify the connection is alive. Open does not
	// verify connectivity (pgxpool connects lazily), so the explicit Ping below
	// is what makes an unreachable database fail startup.
	store, err := postgres.Open(
		ctx,
		postgres.WithURI(ctx, application.cfg.Database.URI),
		postgres.WithPingTimeout(application.cfg.Database.PingTimeout),
		postgres.WithLogger(application.logger),
	)
	if err != nil {
		return fmt.Errorf("postgres.Open(): %w", err)
	}

	if err := store.Ping(ctx); err != nil {
		// The pool is already open at this point; close it so a failed startup
		// does not leak connections or background health-check goroutines.
		if closeErr := store.Close(); closeErr != nil {
			application.logger.Error("error closing database connection after failed ping", zap.Error(closeErr))
		}
		return fmt.Errorf("store.Ping(): %w", err)
	}

	// The pool outlives the component group: it is closed only after every
	// component has returned, so in-flight requests can still reach the
	// database while the webserver drains.
	defer func() {
		if err := store.Close(); err != nil {
			application.logger.Error("error closing database connection", zap.Error(err))
		}
	}()

	// Configure the controllers
	controllerOpts := controller.Options{
		ReadyState: application.ready,
		Store:      store,
	}
	appController, err := controller.NewController(&controllerOpts)
	if err != nil {
		return fmt.Errorf("controller.NewController(): %w", err)
	}

	// Configure the API server
	appServer, err := api.NewServer(appController)
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
		web.WithReadTimeout(application.cfg.Webserver.ReadTimeout),
		web.WithWriteTimeout(application.cfg.Webserver.WriteTimeout),
		web.WithIdleTimeout(application.cfg.Webserver.IdleTimeout),
		web.WithLogger(application.logger),
		web.WithFileServer(staticFs),
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

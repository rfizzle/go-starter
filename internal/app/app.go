package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/internal/api/v1"
	"github.com/rfizzle/go-starter/internal/controller"
	"github.com/rfizzle/go-starter/internal/utils"
	"github.com/rfizzle/go-starter/internal/web"
	"github.com/rfizzle/go-starter/pkg/config"
	staticweb "github.com/rfizzle/go-starter/web"
)

type App struct {
	cfg         *config.Application
	state       *state
	stopProcess []func()
	ready       *atomic.Bool
	logger      *zap.Logger
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

func (application *App) Start() error {
	var wg sync.WaitGroup

	// Configure the controllers
	controllerOpts := controller.Options{
		ReadyState: application.ready,
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

	// Go ahead and create the stop process before starting the server
	application.stopProcess = utils.PrependFunc(application.stopProcess, func() {
		err := webServer.Stop()
		if err != nil {
			application.logger.Error("error stopping webserver", zap.Error(err))
		}
	})

	// Run our server in a goroutine so that it doesn't block.
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := webServer.Start()
		if err != nil {
			application.logger.Error("error starting webserver", zap.Error(err))
		}
	}()

	// Set the ready flag to true
	application.ready.Store(true)

	// Log that we are serving the application
	application.logger.Info(fmt.Sprintf("application started on %s:%d", application.cfg.Webserver.Address, application.cfg.Webserver.Port))
	wg.Wait()

	return nil
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

func (application *App) Stop() error {
	application.logger.Info("stopping app")
	application.ready.Store(false)
	for i := len(application.stopProcess) - 1; i >= 0; i-- {
		application.stopProcess[i]()
	}

	return nil
}

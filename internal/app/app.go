package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rfizzle/go-starter/internal/controller"

	"github.com/rfizzle/go-starter/internal/server"
	"github.com/rfizzle/go-starter/internal/utils"
	"github.com/rfizzle/go-starter/web"

	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/pkg/config"
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

	// Configure logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("zap.NewProduction(): %w", err)
	}

	return &App{
		cfg: cfg,
		state: &state{
			mutex:      sync.Mutex{},
			StartTime:  time.Now(),
			ErrorCount: 0,
		},
		ready:  isReady,
		logger: logger,
	}, nil
}

func (application *App) Start() error {
	var wg sync.WaitGroup

	// Configure the controllers
	appController := controller.NewController(application.ready)

	// Configure static file server
	staticFs := web.AssetHandler("/", "")

	// Configure application webserver
	appServer, err := server.New(
		appController,
		server.WithFileServer(staticFs),
		server.WithAddress(application.cfg.Webserver.Address),
		server.WithPort(application.cfg.Webserver.Port),
		server.WithReadTimeout(application.cfg.Webserver.ReadTimeout),
		server.WithWriteTimeout(application.cfg.Webserver.WriteTimeout),
		server.WithIdleTimeout(application.cfg.Webserver.IdleTimeout),
		server.WithLogger(application.logger),
	)
	if err != nil {
		return fmt.Errorf("server.New(): %w", err)
	}

	// Go ahead and create the stop process before starting the server
	application.stopProcess = utils.PrependFunc(application.stopProcess, func() {
		err := appServer.Stop()
		if err != nil {
			application.logger.Error("error stopping webserver", zap.Error(err))
		}
	})

	// Run our server in a goroutine so that it doesn't block.
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := appServer.Start()
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

func (application *App) Stop() error {
	application.logger.Info("stopping app")
	application.ready.Store(false)
	for i := len(application.stopProcess) - 1; i >= 0; i-- {
		application.stopProcess[i]()
	}

	return nil
}

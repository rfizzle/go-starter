package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/pkg/config"
)

type App struct {
	cfg         *config.Config
	state       *state
	stopProcess []func() error
	ready       *atomic.Value
	logger      *zap.Logger
}

func New(cfg *config.Config) (*App, error) {
	isReady := &atomic.Value{}
	isReady.Store(false)

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

func (a *App) Start() error {
	a.logger.Info("starting app")
	a.ready.Store(true)
	return nil
}

func (a *App) Ready() bool {
	return a.ready.Load().(bool)
}

func (a *App) SetReady(ready bool) {
	a.ready.Store(ready)
}

func (a *App) Logger() *zap.Logger {
	return a.logger
}

func (a *App) Stop() error {
	a.logger.Info("stopping app")
	a.ready.Store(false)
	for i := len(a.stopProcess) - 1; i >= 0; i-- {
		err := a.stopProcess[i]()
		if err != nil {
			a.logger.Error("stop process error", zap.Error(err))
		}
	}

	return nil
}

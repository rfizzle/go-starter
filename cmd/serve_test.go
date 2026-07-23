package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/rfizzle/go-starter/pkg/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type syncErrorWriteSyncer struct {
	err error
}

func (writer syncErrorWriteSyncer) Write(data []byte) (int, error) {
	return len(data), nil
}

func (writer syncErrorWriteSyncer) Sync() error {
	return writer.err
}

type fakeServeApplication struct {
	run      func(context.Context) error
	logger   *zap.Logger
	ready    bool
	runCalls int
}

func validServeCommandConfig() *config.Application {
	return &config.Application{
		Environment: "production",
		Webserver: config.Webserver{
			Port: 8080,
		},
		Database: config.Database{
			URI: "postgres://localhost:5432/example",
		},
		Auth: config.Auth{
			JWTSecret: strings.Repeat("s", 32),
			JWTTTL:    3600,
		},
		Logging: config.Logging{
			Level: "info",
		},
	}
}

func (application *fakeServeApplication) Run(ctx context.Context) error {
	application.runCalls++
	return application.run(ctx)
}

func (application *fakeServeApplication) SetLogger(logger *zap.Logger) {
	application.logger = logger
}

func (application *fakeServeApplication) Logger() *zap.Logger {
	return application.logger
}

func (application *fakeServeApplication) SetReady(ready bool) {
	application.ready = ready
}

func TestServeCommand_PropagatesStartupFailures(t *testing.T) {
	configErr := errors.New("config failed")
	loggerErr := errors.New("logger failed")
	applicationErr := errors.New("application failed")
	databaseErr := errors.New("store.Ping(): connection refused")
	listenerErr := errors.New("issue serving webserver: address already in use")

	tests := []struct {
		name       string
		configErr  error
		loggerErr  error
		appErr     error
		runErr     error
		wantErr    error
		wantPrefix string
	}{
		{
			name:       "configuration failure",
			configErr:  configErr,
			wantErr:    configErr,
			wantPrefix: "resolve config",
		},
		{
			name:       "logger failure",
			loggerErr:  loggerErr,
			wantErr:    loggerErr,
			wantPrefix: "initialize logger",
		},
		{
			name:       "application construction failure",
			appErr:     applicationErr,
			wantErr:    applicationErr,
			wantPrefix: "initialize application",
		},
		{
			name:       "database startup failure",
			runErr:     databaseErr,
			wantErr:    databaseErr,
			wantPrefix: "run application",
		},
		{
			name:       "listener startup failure",
			runErr:     listenerErr,
			wantErr:    listenerErr,
			wantPrefix: "run application",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application := &fakeServeApplication{
				run: func(context.Context) error {
					return tt.runErr
				},
			}
			deps := serveDependencies{
				resolveConfig: func() (*config.Application, error) {
					return validServeCommandConfig(), tt.configErr
				},
				newLogger: func(*config.Application) (*zap.Logger, error) {
					return zap.NewNop(), tt.loggerErr
				},
				newApplication: func(*config.Application) (serveApplication, error) {
					return application, tt.appErr
				},
				syncLogger: func(*zap.Logger) error {
					return nil
				},
			}

			err := newServeCommand(deps).Execute()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Execute() error = %v, want error wrapping %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantPrefix) {
				t.Errorf("Execute() error = %q, want prefix %q", err, tt.wantPrefix)
			}
		})
	}
}

func TestServeCommand_CleanContextShutdownSucceeds(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	cancel()

	application := &fakeServeApplication{
		run: func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		},
	}
	syncCalls := 0
	deps := serveDependencies{
		resolveConfig: func() (*config.Application, error) {
			return validServeCommandConfig(), nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			return zap.NewNop(), nil
		},
		newApplication: func(*config.Application) (serveApplication, error) {
			return application, nil
		},
		syncLogger: func(*zap.Logger) error {
			syncCalls++
			return nil
		},
	}

	command := newServeCommand(deps)
	command.SetContext(parent)
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	if application.runCalls != 1 {
		t.Errorf("application Run() calls = %d, want 1", application.runCalls)
	}
	if syncCalls != 1 {
		t.Errorf("logger Sync() calls = %d, want 1", syncCalls)
	}
}

func TestServeCommand_ValidatesBeforeStartup(t *testing.T) {
	cfg := validServeCommandConfig()
	cfg.Database.URI = ""

	deps := serveDependencies{
		resolveConfig: func() (*config.Application, error) {
			return cfg, nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			t.Fatal("newLogger() called before configuration validation")
			return nil, nil
		},
		newApplication: func(*config.Application) (serveApplication, error) {
			t.Fatal("newApplication() called before configuration validation")
			return nil, nil
		},
		syncLogger: func(*zap.Logger) error {
			t.Fatal("syncLogger() called before logger construction")
			return nil
		},
	}

	err := newServeCommand(deps).Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "validate config: database.uri") {
		t.Errorf("Execute() error = %q, want database URI validation context", err)
	}
}

func TestServeCommand_ConfigFileReadFailureUsesCobraErrorPath(t *testing.T) {
	previousConfigFile := cfgFile
	t.Cleanup(func() {
		cfgFile = previousConfigFile
		viper.Reset()
	})
	cfgFile = filepath.Join(t.TempDir(), "missing.yml")

	deps := serveDependencies{
		resolveConfig: resolveConfig,
		newLogger: func(*config.Application) (*zap.Logger, error) {
			t.Fatal("newLogger() called after config read failure")
			return nil, nil
		},
		newApplication: func(*config.Application) (serveApplication, error) {
			t.Fatal("newApplication() called after config read failure")
			return nil, nil
		},
		syncLogger: func(*zap.Logger) error {
			t.Fatal("syncLogger() called before logger construction")
			return nil
		},
	}

	err := newServeCommand(deps).Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want config file read error")
	}
	if !strings.Contains(err.Error(), "resolve config: read config") {
		t.Errorf("Execute() error = %q, want config read context", err)
	}
}

func TestServeCommand_LoggerCleanupPreservesRunError(t *testing.T) {
	runErr := errors.New("run failed")
	syncErr := errors.New("sync failed")
	application := &fakeServeApplication{
		run: func(context.Context) error {
			return runErr
		},
	}
	deps := serveDependencies{
		resolveConfig: func() (*config.Application, error) {
			return validServeCommandConfig(), nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			return zap.NewNop(), nil
		},
		newApplication: func(*config.Application) (serveApplication, error) {
			return application, nil
		},
		syncLogger: func(*zap.Logger) error {
			return syncErr
		},
	}

	err := newServeCommand(deps).Execute()
	if !errors.Is(err, runErr) {
		t.Errorf("Execute() error = %v, want run error preserved", err)
	}
	if !errors.Is(err, syncErr) {
		t.Errorf("Execute() error = %v, want sync error preserved", err)
	}
}

func TestSyncLogger(t *testing.T) {
	syncErr := errors.New("sync failed")
	tests := []struct {
		name    string
		syncErr error
		wantErr error
	}{
		{
			name:    "invalid terminal operation is harmless",
			syncErr: &os.PathError{Op: "sync", Path: "/dev/stderr", Err: syscall.EINVAL},
		},
		{
			name:    "unsupported terminal ioctl is harmless",
			syncErr: &os.PathError{Op: "sync", Path: "/dev/stderr", Err: syscall.ENOTTY},
		},
		{
			name:    "genuine sync failure propagates",
			syncErr: syncErr,
			wantErr: syncErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				syncErrorWriteSyncer{err: tt.syncErr},
				zap.DebugLevel,
			)

			err := syncLogger(zap.New(core))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("syncLogger() error = %v, want error wrapping %v", err, tt.wantErr)
			}
		})
	}
}

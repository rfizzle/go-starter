package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/rfizzle/go-starter/pkg/config"
	"go.uber.org/zap"
)

func TestSeedCommand_ValidatesBeforeStartup(t *testing.T) {
	cfg := validSeedCommandConfig()
	cfg.Database.URI = ""

	deps := seedDependencies{
		resolveConfig: func() (*config.Application, error) {
			return cfg, nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			t.Fatal("newLogger() called before configuration validation")
			return nil, nil
		},
		run: func(context.Context, *config.Application, *zap.Logger) error {
			t.Fatal("run() called before configuration validation")
			return nil
		},
	}

	err := newSeedCommand(deps).Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "validate config: database.uri") {
		t.Errorf("Execute() error = %q, want database URI validation context", err)
	}
}

func TestSeedCommand_UsesSeedValidationScope(t *testing.T) {
	cfg := validSeedCommandConfig()
	cfg.Webserver.Port = -1
	cfg.Auth.JWTSecret = ""
	cfg.Auth.JWTTTL = -1
	cfg.Observability.Enabled = true
	cfg.Observability.Endpoint = "not-an-endpoint"
	cfg.Observability.SampleRatio = 2
	cfg.Database.CloseTimeout = -1

	runCalls := 0
	deps := seedDependencies{
		resolveConfig: func() (*config.Application, error) {
			return cfg, nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			return zap.NewNop(), nil
		},
		run: func(context.Context, *config.Application, *zap.Logger) error {
			runCalls++
			return nil
		},
	}

	if err := newSeedCommand(deps).Execute(); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if runCalls != 1 {
		t.Errorf("run() calls = %d, want 1", runCalls)
	}
}

func validSeedCommandConfig() *config.Application {
	return &config.Application{
		Environment: "production",
		Database: config.Database{
			URI:         "postgres://localhost:5432/example",
			PingTimeout: 15,
		},
		Logging: config.Logging{
			Level: "info",
		},
	}
}

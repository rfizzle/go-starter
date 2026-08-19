package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rfizzle/go-starter/internal/cli/seed"
	"github.com/rfizzle/go-starter/internal/platform/config"
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
		run: func(context.Context, *config.Application, *zap.Logger, seed.UserSeed) (seed.Result, error) {
			t.Fatal("run() called before configuration validation")
			return seed.Result{}, nil
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

func TestSeedCommand_RefusesProductionEnvironment(t *testing.T) {
	cfg := validSeedCommandConfig()
	cfg.Environment = "production"
	cfg.Seed.Password = "explicit-password"

	deps := seedDependencies{
		resolveConfig: func() (*config.Application, error) {
			return cfg, nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			t.Fatal("newLogger() called for production seed")
			return nil, nil
		},
		run: func(context.Context, *config.Application, *zap.Logger, seed.UserSeed) (seed.Result, error) {
			t.Fatal("run() called for production seed")
			return seed.Result{}, nil
		},
	}

	err := newSeedCommand(deps).Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want production environment guard")
	}
	if !strings.Contains(err.Error(), "only supported in the development environment") {
		t.Errorf("Execute() error = %q, want development-only context", err)
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
		run: func(context.Context, *config.Application, *zap.Logger, seed.UserSeed) (seed.Result, error) {
			runCalls++
			return seed.Result{}, nil
		},
	}

	if err := newSeedCommand(deps).Execute(); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if runCalls != 1 {
		t.Errorf("run() calls = %d, want 1", runCalls)
	}
}

func TestSeedCommand_DoesNotDiscloseConfiguredPassword(t *testing.T) {
	cfg := validSeedCommandConfig()

	deps := seedDependencies{
		resolveConfig: func() (*config.Application, error) {
			return cfg, nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			return zap.NewNop(), nil
		},
		run: func(
			_ context.Context,
			_ *config.Application,
			_ *zap.Logger,
			userSeed seed.UserSeed,
		) (seed.Result, error) {
			if userSeed.Password != cfg.Seed.Password {
				t.Errorf("run seed password = %q, want configured password", userSeed.Password)
			}
			return seed.Result{Created: 1}, nil
		},
	}

	output := new(bytes.Buffer)
	command := newSeedCommand(deps)
	command.SetOut(output)
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if strings.Contains(output.String(), cfg.Seed.Password) {
		t.Errorf("output exposed configured password: %q", output.String())
	}
}

func validSeedCommandConfig() *config.Application {
	return &config.Application{
		Environment: "development",
		Database: config.Database{
			URI:         "postgres://localhost:5432/example",
			PingTimeout: 15,
		},
		Seed: config.Seed{
			Email:     "developer@example.com",
			Password:  "development-password",
			FirstName: "Development",
			LastName:  "User",
		},
		Logging: config.Logging{
			Level: "info",
		},
	}
}

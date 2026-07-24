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
		generatePassword: func() (string, error) {
			t.Fatal("generatePassword() called before configuration validation")
			return "", nil
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
		generatePassword: func() (string, error) {
			t.Fatal("generatePassword() called for production seed")
			return "", nil
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
		generatePassword: func() (string, error) {
			return "generated-password", nil
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

func TestSeedCommand_DisclosesGeneratedPasswordOnlyForCreatedUser(t *testing.T) {
	tests := []struct {
		name         string
		created      int
		wantPassword bool
	}{
		{name: "created user", created: 1, wantPassword: true},
		{name: "existing user", created: 0, wantPassword: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validSeedCommandConfig()
			const generatedPassword = "one-time-generated-password"
			var receivedSeed seed.UserSeed

			deps := seedDependencies{
				resolveConfig: func() (*config.Application, error) {
					return cfg, nil
				},
				newLogger: func(*config.Application) (*zap.Logger, error) {
					return zap.NewNop(), nil
				},
				generatePassword: func() (string, error) {
					return generatedPassword, nil
				},
				run: func(
					_ context.Context,
					_ *config.Application,
					_ *zap.Logger,
					userSeed seed.UserSeed,
				) (seed.Result, error) {
					receivedSeed = userSeed
					return seed.Result{Created: tt.created, Existing: 1 - tt.created}, nil
				},
			}

			output := new(bytes.Buffer)
			command := newSeedCommand(deps)
			command.SetOut(output)
			if err := command.Execute(); err != nil {
				t.Fatalf("Execute() unexpected error: %v", err)
			}

			if receivedSeed.Password != generatedPassword {
				t.Errorf("run seed password = %q, want generated password", receivedSeed.Password)
			}
			gotCount := strings.Count(output.String(), generatedPassword)
			if tt.wantPassword && gotCount != 1 {
				t.Errorf("generated password output count = %d, want 1; output = %q", gotCount, output.String())
			}
			if !tt.wantPassword && gotCount != 0 {
				t.Errorf("generated password output count = %d, want 0; output = %q", gotCount, output.String())
			}
		})
	}
}

func TestSeedCommand_DoesNotDiscloseConfiguredPassword(t *testing.T) {
	cfg := validSeedCommandConfig()
	cfg.Seed.Password = "configured-private-password"

	deps := seedDependencies{
		resolveConfig: func() (*config.Application, error) {
			return cfg, nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			return zap.NewNop(), nil
		},
		generatePassword: func() (string, error) {
			t.Fatal("generatePassword() called for configured password")
			return "", nil
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
			FirstName: "Development",
			LastName:  "User",
		},
		Logging: config.Logging{
			Level: "info",
		},
	}
}

package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/internal/platform/config"
	"github.com/rfizzle/go-starter/internal/platform/database"
)

func TestMigrateCommand_Actions(t *testing.T) {
	tests := []struct {
		name       string
		action     migrationAction
		status     database.MigrationStatus
		wantOutput string
	}{
		{
			name:       "up applies migrations",
			action:     migrationUp,
			status:     database.StatusComplete,
			wantOutput: "database migrations complete\n",
		},
		{
			name:       "status reports none",
			action:     migrationStatus,
			status:     database.StatusNone,
			wantOutput: "migration status: none\n",
		},
		{
			name:       "status reports partial",
			action:     migrationStatus,
			status:     database.StatusPartial,
			wantOutput: "migration status: partial\n",
		},
		{
			name:       "status reports complete",
			action:     migrationStatus,
			status:     database.StatusComplete,
			wantOutput: "migration status: complete\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAction migrationAction
			deps := migrateDependencies{
				resolveConfig: func() (*config.Application, error) {
					return validMigrateCommandConfig(), nil
				},
				newLogger: func(*config.Application) (*zap.Logger, error) {
					return zap.NewNop(), nil
				},
				run: func(
					ctx context.Context,
					cfg *config.Application,
					appLogger *zap.Logger,
					action migrationAction,
				) (database.MigrationStatus, error) {
					gotAction = action
					return tt.status, nil
				},
			}

			var output bytes.Buffer
			command := newMigrateCommand(deps)
			command.SetOut(&output)
			command.SetArgs([]string{string(tt.action)})

			if err := command.Execute(); err != nil {
				t.Fatalf("Execute() error = %v, want nil", err)
			}
			if gotAction != tt.action {
				t.Errorf("run() action = %q, want %q", gotAction, tt.action)
			}
			if output.String() != tt.wantOutput {
				t.Errorf("output = %q, want %q", output.String(), tt.wantOutput)
			}
		})
	}
}

func TestMigrateCommand_ValidatesBeforeDatabaseWork(t *testing.T) {
	cfg := validMigrateCommandConfig()
	cfg.Database.URI = ""

	deps := migrateDependencies{
		resolveConfig: func() (*config.Application, error) {
			return cfg, nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			t.Fatal("newLogger() called before configuration validation")
			return nil, nil
		},
		run: func(
			context.Context,
			*config.Application,
			*zap.Logger,
			migrationAction,
		) (database.MigrationStatus, error) {
			t.Fatal("run() called before configuration validation")
			return database.StatusNone, nil
		},
	}

	command := newMigrateCommand(deps)
	command.SetArgs([]string{"up"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "validate config: database.uri") {
		t.Fatalf("Execute() error = %v, want database URI validation error", err)
	}
}

func TestMigrateCommand_PropagatesMigrationFailure(t *testing.T) {
	migrationErr := errors.New("database unavailable")
	deps := migrateDependencies{
		resolveConfig: func() (*config.Application, error) {
			return validMigrateCommandConfig(), nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			return zap.NewNop(), nil
		},
		run: func(
			context.Context,
			*config.Application,
			*zap.Logger,
			migrationAction,
		) (database.MigrationStatus, error) {
			return database.StatusNone, migrationErr
		},
	}

	command := newMigrateCommand(deps)
	command.SetArgs([]string{"up"})
	err := command.Execute()
	if !errors.Is(err, migrationErr) {
		t.Fatalf("Execute() error = %v, want error wrapping %v", err, migrationErr)
	}
}

func TestMigrateCommand_PropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	deps := migrateDependencies{
		resolveConfig: func() (*config.Application, error) {
			return validMigrateCommandConfig(), nil
		},
		newLogger: func(*config.Application) (*zap.Logger, error) {
			return zap.NewNop(), nil
		},
		run: func(
			ctx context.Context,
			cfg *config.Application,
			appLogger *zap.Logger,
			action migrationAction,
		) (database.MigrationStatus, error) {
			return database.StatusNone, ctx.Err()
		},
	}

	command := newMigrateCommand(deps)
	command.SetContext(ctx)
	command.SetArgs([]string{"up"})
	err := command.Execute()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute() error = %v, want context cancellation", err)
	}
}

func TestServeCommand_AutoMigrationDefaultsDisabled(t *testing.T) {
	flag := newServeCommand(defaultServeDependencies).Flags().Lookup("database.auto-migrate")
	if flag == nil {
		t.Fatal("database.auto-migrate flag is not registered")
	}
	if flag.DefValue != "false" {
		t.Errorf("database.auto-migrate default = %q, want false", flag.DefValue)
	}
}

func validMigrateCommandConfig() *config.Application {
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

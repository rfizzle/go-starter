package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/internal/platform/config"
	"github.com/rfizzle/go-starter/internal/platform/database"
	"github.com/rfizzle/go-starter/internal/platform/database/postgres"
	"github.com/rfizzle/go-starter/internal/platform/observability"
)

type migrationAction string

const (
	migrationUp     migrationAction = "up"
	migrationStatus migrationAction = "status"
)

type migrateDependencies struct {
	resolveConfig func() (*config.Application, error)
	newLogger     func(*config.Application) (*zap.Logger, error)
	run           func(context.Context, *config.Application, *zap.Logger, migrationAction) (database.MigrationStatus, error)
}

var defaultMigrateDependencies = migrateDependencies{
	resolveConfig: resolveConfig,
	newLogger:     observability.NewLogger,
	run:           runMigration,
}

var migrateCmd = newMigrateCommand(defaultMigrateDependencies)

func newMigrateCommand(deps migrateDependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "migrate",
		Short: "Manage embedded PostgreSQL migrations.",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(
		newMigrationActionCommand(deps, migrationUp),
		newMigrationActionCommand(deps, migrationStatus),
	)
	return command
}

func newMigrationActionCommand(deps migrateDependencies, action migrationAction) *cobra.Command {
	command := &cobra.Command{
		Use:   string(action),
		Short: migrationActionDescription(action),
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return fmt.Errorf("bind migrate flags: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := deps.resolveConfig()
			if err != nil {
				return fmt.Errorf("resolve config: %w", err)
			}
			if err := cfg.Validate(config.ValidationScopeMigrate); err != nil {
				return fmt.Errorf("validate config: %w", err)
			}

			appLogger, err := deps.newLogger(cfg)
			if err != nil {
				return fmt.Errorf("initialize logger: %w", err)
			}
			logger = appLogger

			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			status, err := deps.run(ctx, cfg, appLogger, action)
			if err != nil {
				return fmt.Errorf("migrate %s: %w", action, err)
			}

			switch action {
			case migrationUp:
				_, err = fmt.Fprintln(cmd.OutOrStdout(), "database migrations complete")
			case migrationStatus:
				var label string
				label, err = migrationStatusLabel(status)
				if err == nil {
					_, err = fmt.Fprintf(cmd.OutOrStdout(), "migration status: %s\n", label)
				}
			}
			if err != nil {
				return fmt.Errorf("write migrate output: %w", err)
			}
			return nil
		},
	}
	addMigrationFlags(command)
	return command
}

func migrationActionDescription(action migrationAction) string {
	switch action {
	case migrationUp:
		return "Apply all pending forward-only migrations."
	case migrationStatus:
		return "Report whether embedded migrations are applied."
	default:
		return ""
	}
}

func addMigrationFlags(command *cobra.Command) {
	command.Flags().String("environment", "production", "runtime environment (development, staging, production); selects logger defaults")
	addDatabaseFlags(command)
	addLoggingFlags(command)
}

func runMigration(
	ctx context.Context,
	cfg *config.Application,
	appLogger *zap.Logger,
	action migrationAction,
) (database.MigrationStatus, error) {
	pool, err := openDatabase(ctx, cfg.Database)
	if err != nil {
		return database.StatusNone, err
	}
	defer pool.Close()

	switch action {
	case migrationUp:
		if err := postgres.Migrate(ctx, pool, appLogger, postgres.Migrations); err != nil {
			return database.StatusNone, fmt.Errorf("postgres.Migrate(): %w", err)
		}
		return database.StatusComplete, nil
	case migrationStatus:
		status, err := postgres.MigrationStatus(ctx, pool, postgres.Migrations)
		if err != nil {
			return database.StatusNone, fmt.Errorf("postgres.MigrationStatus(): %w", err)
		}
		return status, nil
	default:
		return database.StatusNone, fmt.Errorf("unsupported migration action %q", action)
	}
}

func migrationStatusLabel(status database.MigrationStatus) (string, error) {
	switch status {
	case database.StatusNone:
		return "none", nil
	case database.StatusPartial:
		return "partial", nil
	case database.StatusComplete:
		return "complete", nil
	default:
		return "", fmt.Errorf("unsupported migration status %d", status)
	}
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

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
	command.Flags().String("database.uri", "", "PostgreSQL connection URI (e.g. postgres://user:pass@host:5432/db)")
	command.Flags().Int("database.ping-timeout", 15, "timeout in seconds for database ping operations")
	command.Flags().Int("database.max-conns", 0, "maximum size of the connection pool (0 uses the pgx default: max(4, NumCPU))")
	command.Flags().Int("database.min-conns", 0, "number of idle connections kept warm in the pool (0 uses the pgx default: 0)")
	command.Flags().Int("database.max-conn-lifetime", 0, "seconds after which a connection is closed and replaced (0 uses the pgx default: 3600)")
	command.Flags().Int("database.max-conn-idle-time", 0, "seconds after which an idle connection is closed (0 uses the pgx default: 1800)")
	command.Flags().Int("database.health-check-period", 0, "seconds between pool health checks of idle connections (0 uses the pgx default: 60)")
	command.Flags().String("logging.level", "info", "minimum log level: debug, info, warn, or error (--verbose forces debug)")
	command.Flags().String("logging.format", "", "log encoding: json or console (defaults to console in development, json otherwise)")
}

func runMigration(
	ctx context.Context,
	cfg *config.Application,
	appLogger *zap.Logger,
	action migrationAction,
) (database.MigrationStatus, error) {
	pool, err := postgres.NewPool(ctx, cfg.Database.URI, postgres.PoolConfig{
		MaxConns:                 cfg.Database.MaxConns,
		MinConns:                 cfg.Database.MinConns,
		MaxConnLifetimeSeconds:   cfg.Database.MaxConnLifetime,
		MaxConnIdleTimeSeconds:   cfg.Database.MaxConnIdleTime,
		HealthCheckPeriodSeconds: cfg.Database.HealthCheckPeriod,
	})
	if err != nil {
		return database.StatusNone, fmt.Errorf("postgres.NewPool(): %w", err)
	}
	defer pool.Close()

	store, err := postgres.Open(
		ctx,
		postgres.WithPool(pool),
		postgres.WithPingTimeout(cfg.Database.PingTimeout),
		postgres.WithLogger(appLogger),
	)
	if err != nil {
		return database.StatusNone, fmt.Errorf("postgres.Open(): %w", err)
	}
	if err := store.Ping(ctx); err != nil {
		return database.StatusNone, fmt.Errorf("store.Ping(): %w", err)
	}

	switch action {
	case migrationUp:
		if err := store.Migrate(ctx, postgres.Migrations); err != nil {
			return database.StatusNone, fmt.Errorf("store.Migrate(): %w", err)
		}
		return database.StatusComplete, nil
	case migrationStatus:
		status, err := store.MigrationStatus(ctx, postgres.Migrations)
		if err != nil {
			return database.StatusNone, fmt.Errorf("store.MigrationStatus(): %w", err)
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

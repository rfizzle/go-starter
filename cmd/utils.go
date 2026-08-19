package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/internal/platform/config"
	"github.com/rfizzle/go-starter/internal/platform/database/postgres"
)

// resolveConfig unmarshals every viper layer (flag > env > file > default) into
// a config.Application, stamps the build metadata that lives in ldflags rather
// than config, and applies the --verbose debug shortcut. It is the shared
// front-half of every long-running subcommand (serve, seed) so config handling
// stays identical across them.
func resolveConfig() (*config.Application, error) {
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if cfgFile != "" || !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := &config.Application{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Build-time metadata is stamped via ldflags, not config layers, so set it
	// directly after unmarshaling.
	cfg.Build = config.Build{
		Branch:   BuildBranch,
		Date:     BuildDate,
		Env:      BuildEnv,
		Revision: BuildRevision,
		Version:  BuildVersion,
	}

	// The --verbose flag is a shortcut for debug logging and outranks the
	// configured level.
	if verbose {
		cfg.Logging.Level = "debug"
	}

	return cfg, nil
}

// persistentPreRunEHook chains a command's PersistentPreRunE onto its parent's,
// which cobra otherwise replaces rather than composes.
func persistentPreRunEHook(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.HasParent() {
			parent := cmd.Parent()
			if parent.PersistentPreRunE != nil {
				if err := parent.PersistentPreRunE(parent, args); err != nil {
					return err
				}
			} else if parent.PersistentPreRun != nil {
				parent.PersistentPreRun(parent, args)
			}
		}
		if fn != nil {
			return fn(cmd, args)
		}
		return nil
	}
}

// addDatabaseFlags registers the connection-pool flags every database-backed
// subcommand (serve, seed, migrate) accepts, so their names, defaults, and help
// text cannot drift between commands. Flags specific to one command (for
// example serve's auto-migrate) stay with that command.
func addDatabaseFlags(command *cobra.Command) {
	command.Flags().String("database.uri", "", "PostgreSQL connection URI (e.g. postgres://user:pass@host:5432/db)")
	command.Flags().Int("database.ping-timeout", 15, "timeout in seconds for database ping operations")
	command.Flags().Int("database.max-conns", 0, "maximum size of the connection pool (0 uses the pgx default: max(4, NumCPU))")
	command.Flags().Int("database.min-conns", 0, "number of idle connections kept warm in the pool (0 uses the pgx default: 0)")
	command.Flags().Int("database.max-conn-lifetime", 0, "seconds after which a connection is closed and replaced (0 uses the pgx default: 3600)")
	command.Flags().Int("database.max-conn-idle-time", 0, "seconds after which an idle connection is closed (0 uses the pgx default: 1800)")
	command.Flags().Int("database.health-check-period", 0, "seconds between pool health checks of idle connections (0 uses the pgx default: 60)")
}

// addLoggingFlags registers the logger flags shared by serve, seed, and migrate
// so each command honours the same flags rather than only config file / env vars.
func addLoggingFlags(command *cobra.Command) {
	command.Flags().String("logging.level", "info", "minimum log level: debug, info, warn, or error (--verbose forces debug)")
	command.Flags().String("logging.format", "", "log encoding: json or console (defaults to console in development, json otherwise)")
}

// openDatabase builds the connection pool from config, wraps it in the
// lifecycle client, and pings it so an unreachable database fails before any
// work starts. It is the shared front-half of the short-lived database
// commands (seed, migrate); serve owns a longer pool lifecycle in internal/app.
// The caller owns the returned pool and must close it; the client borrows it.
func openDatabase(ctx context.Context, cfg config.Database, appLogger *zap.Logger) (*pgxpool.Pool, *postgres.Client, error) {
	pool, err := postgres.NewPool(ctx, cfg.URI, postgres.PoolConfigFromDatabase(cfg))
	if err != nil {
		return nil, nil, fmt.Errorf("postgres.NewPool(): %w", err)
	}

	store, err := postgres.Open(
		ctx,
		postgres.WithPool(pool),
		postgres.WithPingTimeout(cfg.PingTimeout),
		postgres.WithLogger(appLogger),
	)
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("postgres.Open(): %w", err)
	}

	if err := store.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("store.Ping(): %w", err)
	}

	return pool, store, nil
}

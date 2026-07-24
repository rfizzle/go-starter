package cmd

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/internal/cli/seed"
	"github.com/rfizzle/go-starter/internal/modules/user"
	"github.com/rfizzle/go-starter/internal/platform/config"
	"github.com/rfizzle/go-starter/internal/platform/database/postgres"
	"github.com/rfizzle/go-starter/internal/platform/observability"
)

type seedDependencies struct {
	resolveConfig    func() (*config.Application, error)
	newLogger        func(*config.Application) (*zap.Logger, error)
	generatePassword func() (string, error)
	run              func(context.Context, *config.Application, *zap.Logger, seed.UserSeed) (seed.Result, error)
}

var defaultSeedDependencies = seedDependencies{
	resolveConfig:    resolveConfig,
	newLogger:        observability.NewLogger,
	generatePassword: generateSeedPassword,
	run:              runSeed,
}

// seedCmd inserts the baseline records (see internal/cli/seed) into an already
// migrated database. Seeding is idempotent, so it is safe to run repeatedly.
var seedCmd = newSeedCommand(defaultSeedDependencies)

func newSeedCommand(deps seedDependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "seed",
		Short: "Seed the database with baseline records.",
		Long: `Seed the database with one explicitly configured development user.

The command is refused outside the development environment. Set seed.email and
optionally seed.password. When the password is omitted, a secure random password
is displayed once if the user is created. Existing users are never overwritten.

Example Command:
go-starter seed --environment development --database.uri "$DATABASE_URI" --seed.email admin@example.com`,
		Args:              cobra.NoArgs,
		PersistentPreRunE: persistentPreRunEHook(nil),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return fmt.Errorf("bind seed flags: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := deps.resolveConfig()
			if err != nil {
				return fmt.Errorf("resolve config: %w", err)
			}
			if err := cfg.Validate(config.ValidationScopeSeed); err != nil {
				return fmt.Errorf("validate config: %w", err)
			}

			userSeed := seed.UserSeed{
				Email:     cfg.Seed.Email,
				Password:  cfg.Seed.Password,
				FirstName: cfg.Seed.FirstName,
				LastName:  cfg.Seed.LastName,
			}
			generatedPassword := userSeed.Password == ""
			if generatedPassword {
				userSeed.Password, err = deps.generatePassword()
				if err != nil {
					return fmt.Errorf("generate seed password: %w", err)
				}
			}

			// Replace the bootstrap logger with one driven by config, matching
			// serve, only after validation has succeeded.
			appLogger, err := deps.newLogger(cfg)
			if err != nil {
				return fmt.Errorf("initialize logger: %w", err)
			}
			logger = appLogger

			// Cancel seeding on Ctrl-C or SIGTERM so a long run against a
			// remote database can be interrupted cleanly.
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			result, err := deps.run(ctx, cfg, appLogger, userSeed)
			if err != nil {
				return err
			}
			if generatedPassword && result.Created > 0 {
				if _, err := fmt.Fprintf(
					cmd.OutOrStdout(),
					"Generated seed password for %s: %s\nThis password will not be shown again.\n",
					userSeed.Email,
					userSeed.Password,
				); err != nil {
					return fmt.Errorf("write generated seed password: %w", err)
				}
			}
			return nil
		},
	}

	addSeedFlags(command)
	return command
}

func init() {
	rootCmd.AddCommand(seedCmd)
}

func addSeedFlags(command *cobra.Command) {
	// Application options
	command.Flags().String("environment", "production", "runtime environment (development, staging, production); selects logger defaults")

	// Development seed options.
	command.Flags().String("seed.email", "", "email address for the development seed user (required)")
	command.Flags().String("seed.password", "", "password for the development seed user (empty generates a random password)")
	command.Flags().String("seed.first-name", "", "first name for the development seed user")
	command.Flags().String("seed.last-name", "", "last name for the development seed user")

	// Database options (mirrors serve so seed is runnable standalone).
	command.Flags().String("database.uri", "", "PostgreSQL connection URI, e.g. the value of $DATABASE_URI")
	command.Flags().Int("database.ping-timeout", 15, "timeout in seconds for database ping operations")
	command.Flags().Int("database.max-conns", 0, "maximum size of the connection pool (0 uses the pgx default: max(4, NumCPU))")
	command.Flags().Int("database.min-conns", 0, "number of idle connections kept warm in the pool (0 uses the pgx default: 0)")
	command.Flags().Int("database.max-conn-lifetime", 0, "seconds after which a connection is closed and replaced (0 uses the pgx default: 3600)")
	command.Flags().Int("database.max-conn-idle-time", 0, "seconds after which an idle connection is closed (0 uses the pgx default: 1800)")
	command.Flags().Int("database.health-check-period", 0, "seconds between pool health checks of idle connections (0 uses the pgx default: 60)")

	// Logging options (mirrors serve so seed honours the same flags rather than
	// only config file / env vars).
	command.Flags().String("logging.level", "info", "minimum log level: debug, info, warn, or error (--verbose forces debug)")
	command.Flags().String("logging.format", "", "log encoding: json or console (defaults to console in development, json otherwise)")
}

func generateSeedPassword() (string, error) {
	random := make([]byte, 24)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("read cryptographic randomness: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(random), nil
}

func runSeed(
	ctx context.Context,
	cfg *config.Application,
	appLogger *zap.Logger,
	userSeed seed.UserSeed,
) (seed.Result, error) {
	// Own the pool locally: unlike serve, seed is a short-lived command that
	// connects, seeds, and exits, so it opens and closes its own pool rather
	// than sharing one with long-lived components.
	pool, err := postgres.NewPool(ctx, cfg.Database.URI, postgres.PoolConfig{
		MaxConns:                 cfg.Database.MaxConns,
		MinConns:                 cfg.Database.MinConns,
		MaxConnLifetimeSeconds:   cfg.Database.MaxConnLifetime,
		MaxConnIdleTimeSeconds:   cfg.Database.MaxConnIdleTime,
		HealthCheckPeriodSeconds: cfg.Database.HealthCheckPeriod,
	})
	if err != nil {
		return seed.Result{}, fmt.Errorf("postgres.NewPool(): %w", err)
	}
	defer pool.Close()

	store, err := postgres.Open(
		ctx,
		postgres.WithPool(pool),
		postgres.WithPingTimeout(cfg.Database.PingTimeout),
		postgres.WithLogger(appLogger),
	)
	if err != nil {
		return seed.Result{}, fmt.Errorf("postgres.Open(): %w", err)
	}

	if err := store.Ping(ctx); err != nil {
		return seed.Result{}, fmt.Errorf("store.Ping(): %w", err)
	}

	userRepo := user.NewRepository(pool)

	seeder, err := seed.New(userRepo, []seed.UserSeed{userSeed}, seed.WithLogger(appLogger))
	if err != nil {
		return seed.Result{}, fmt.Errorf("seed.New(): %w", err)
	}

	result, err := seeder.Run(ctx)
	if err != nil {
		return result, fmt.Errorf("seeder.Run(): %w", err)
	}

	appLogger.Info("database seeding complete")
	return result, nil
}

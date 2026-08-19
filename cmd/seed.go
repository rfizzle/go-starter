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

	"github.com/rfizzle/go-starter/internal/cli/seed"
	"github.com/rfizzle/go-starter/internal/modules/user"
	"github.com/rfizzle/go-starter/internal/platform/config"
	"github.com/rfizzle/go-starter/internal/platform/observability"
)

type seedDependencies struct {
	resolveConfig func() (*config.Application, error)
	newLogger     func(*config.Application) (*zap.Logger, error)
	run           func(context.Context, *config.Application, *zap.Logger, seed.UserSeed) (seed.Result, error)
}

var defaultSeedDependencies = seedDependencies{
	resolveConfig: resolveConfig,
	newLogger:     observability.NewLogger,
	run:           runSeed,
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
seed.password; the password is never printed. Existing users are never
overwritten.

Example Command:
go-starter seed --environment development --database.uri "$DATABASE_URI" --seed.email admin@example.com --seed.password "$SEED_PASSWORD"`,
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

			if _, err := deps.run(ctx, cfg, appLogger, userSeed); err != nil {
				return err
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
	command.Flags().String("seed.password", "", "password for the development seed user (required; never printed)")
	command.Flags().String("seed.first-name", "", "first name for the development seed user")
	command.Flags().String("seed.last-name", "", "last name for the development seed user")

	// Database and logging options (mirrors serve so seed is runnable standalone).
	addDatabaseFlags(command)
	addLoggingFlags(command)
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
	pool, err := openDatabase(ctx, cfg.Database)
	if err != nil {
		return seed.Result{}, err
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)

	seeder, err := seed.New(userRepo, appLogger, []seed.UserSeed{userSeed})
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

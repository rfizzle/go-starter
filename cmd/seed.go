package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rfizzle/go-starter/internal/seed"
	"github.com/rfizzle/go-starter/internal/storage/postgres"
	"github.com/rfizzle/go-starter/internal/user"
	"github.com/rfizzle/go-starter/pkg/config"
)

// seedCmd represents the seed command. It applies pending migrations and then
// inserts the baseline records (see internal/seed) into the configured
// database. Seeding is idempotent, so it is safe to run repeatedly.
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database with baseline records.",
	Long: `Seed the database with baseline records.

Applies any pending migrations, then inserts the baseline dataset (an example
user). Seeding is idempotent: records that already exist are skipped, so it is
safe to run more than once.

Example Command:
go-starter seed --database.uri "$DATABASE_URI"`,
	PersistentPreRun: persistentPreRunHook(nil),
	PreRun: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(viper.BindPFlags(cmd.Flags()))
	},
	// Unlike serve (which uses Run + logger.Fatal), seed uses RunE so a failure
	// exits via cobra's error path (non-zero status, error on stderr) — the
	// expected behaviour for a one-shot command invoked from scripts/CI.
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := resolveConfig()
		if err != nil {
			return fmt.Errorf("resolve config: %w", err)
		}

		// Replace the bootstrap logger with one driven by config, matching serve.
		appLogger, err := config.NewLogger(cfg)
		if err != nil {
			return fmt.Errorf("initialize logger: %w", err)
		}
		logger = appLogger

		if cfg.Database.URI == "" {
			return fmt.Errorf("database.uri is required to seed")
		}

		// Cancel migration/seeding on Ctrl-C or SIGTERM so a long run against a
		// remote database can be interrupted cleanly rather than orphaning work.
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// Own the pool locally: unlike serve, seed is a short-lived command that
		// connects, seeds, and exits, so it opens and closes its own pool rather
		// than sharing one with long-lived components.
		pool, err := pgxpool.New(ctx, cfg.Database.URI)
		if err != nil {
			return fmt.Errorf("pgxpool.New(): %w", err)
		}
		defer pool.Close()

		store, err := postgres.Open(
			ctx,
			postgres.WithPool(pool),
			postgres.WithPingTimeout(cfg.Database.PingTimeout),
			postgres.WithLogger(logger),
		)
		if err != nil {
			return fmt.Errorf("postgres.Open(): %w", err)
		}

		if err := store.Ping(ctx); err != nil {
			return fmt.Errorf("store.Ping(): %w", err)
		}

		// Apply migrations first so seeding always runs against the current
		// schema, exactly as serve does before serving traffic.
		if err := store.Migrate(postgres.Migrations); err != nil {
			return fmt.Errorf("store.Migrate(): %w", err)
		}

		userRepo := user.NewRepository(pool)

		seeder, err := seed.New(userRepo, seed.WithLogger(logger))
		if err != nil {
			return fmt.Errorf("seed.New(): %w", err)
		}

		if err := seeder.Run(ctx); err != nil {
			return fmt.Errorf("seeder.Run(): %w", err)
		}

		logger.Info("database seeding complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)

	// Application options
	seedCmd.Flags().String("environment", "production", "runtime environment (development, staging, production); selects logger defaults")

	// Database options (mirrors serve so seed is runnable standalone).
	seedCmd.Flags().String("database.uri", "", "PostgreSQL connection URI, e.g. the value of $DATABASE_URI")
	seedCmd.Flags().Int("database.ping-timeout", 15, "timeout in seconds for database ping operations")

	// Logging options (mirrors serve so seed honours the same flags rather than
	// only config file / env vars).
	seedCmd.Flags().String("logging.level", "info", "minimum log level: debug, info, warn, or error (--verbose forces debug)")
	seedCmd.Flags().String("logging.format", "", "log encoding: json or console (defaults to console in development, json otherwise)")
}

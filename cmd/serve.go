package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rfizzle/go-starter/internal/app"
	"github.com/rfizzle/go-starter/internal/cli"
	"github.com/rfizzle/go-starter/internal/platform/config"
	"github.com/rfizzle/go-starter/internal/platform/observability"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type serveApplication interface {
	Run(context.Context) error
	SetLogger(*zap.Logger)
	Logger() *zap.Logger
	SetReady(bool)
}

type serveDependencies struct {
	resolveConfig  func() (*config.Application, error)
	newLogger      func(*config.Application) (*zap.Logger, error)
	newApplication func(*config.Application) (serveApplication, error)
	syncLogger     func(*zap.Logger) error
}

var defaultServeDependencies = serveDependencies{
	resolveConfig: resolveConfig,
	newLogger:     observability.NewLogger,
	newApplication: func(cfg *config.Application) (serveApplication, error) {
		return app.New(cfg)
	},
	syncLogger: syncLogger,
}

var serveCmd = newServeCommand(defaultServeDependencies)

func newServeCommand(deps serveDependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "serve",
		Short: "Run the go starter application.",
		Long: `Run the go starter application.

Example Command:
go-starter serve --config /etc/crypt/config.json`,
		Args:              cobra.NoArgs,
		PersistentPreRunE: persistentPreRunEHook(nil),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Bind the whole flag set in one call so every serve flag is reachable
			// through viper without a line per key. Flag names use dashes and dots
			// (e.g. "webserver.port") so they map directly onto the nested config
			// struct's mapstructure tags on Unmarshal.
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return fmt.Errorf("bind serve flags: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (runErr error) {
			// Resolve config from all viper layers (flag > env > file > default)
			// into the struct in one pass instead of field-by-field GetString/GetInt.
			cfg, err := deps.resolveConfig()
			if err != nil {
				return fmt.Errorf("resolve config: %w", err)
			}
			if err := cfg.Validate(config.ValidationScopeServe); err != nil {
				return fmt.Errorf("validate config: %w", err)
			}

			// Replace the bootstrap logger with one driven by config (level,
			// format, environment) now that config is resolved.
			appLogger, err := deps.newLogger(cfg)
			if err != nil {
				return fmt.Errorf("initialize logger: %w", err)
			}
			logger = appLogger
			defer func() {
				if err := deps.syncLogger(appLogger); err != nil {
					runErr = errors.Join(runErr, fmt.Errorf("sync logger: %w", err))
				}
			}()

			application, err := deps.newApplication(cfg)
			if err != nil {
				return fmt.Errorf("initialize application: %w", err)
			}
			application.SetLogger(appLogger)

			// Root context cancelled on SIGINT/SIGTERM. WithGracefulShutdown
			// layers the readiness flip and configurable drain delay on top
			// before the app's components are asked to stop.
			signalCtx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			runCtx, cancel := cli.WithGracefulShutdown(
				signalCtx,
				application,
				time.Duration(cfg.ShutdownDelay)*time.Second,
			)
			defer cancel()

			// Run blocks until shutdown or a component fails. Returning the
			// error lets Cobra propagate a nonzero process status.
			if err := application.Run(runCtx); err != nil {
				return fmt.Errorf("run application: %w", err)
			}
			return nil
		},
	}

	addServeFlags(command)
	return command
}

func syncLogger(appLogger *zap.Logger) error {
	err := appLogger.Sync()
	// Syncing stdout/stderr commonly returns EINVAL or ENOTTY when the
	// destination is an unbuffered terminal. All log entries have already been
	// written in those cases, so neither is a cleanup failure.
	if errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY) {
		return nil
	}
	return err
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func addServeFlags(command *cobra.Command) {
	// Application options
	command.Flags().String("environment", "production", "runtime environment (development, staging, production); selects logger defaults")

	// Webserver options
	command.Flags().String("webserver.address", "0.0.0.0", "interface to which the webserver will bind")
	command.Flags().Int("webserver.port", 8080, "port on which the webserver will listen")
	command.Flags().Int("webserver.read-header-timeout", 5, "seconds allowed for reading request headers before timing out (mitigates Slowloris; 0 keeps this server's safe 5s default)")
	command.Flags().Int("webserver.read-timeout", 60, "timeout for webserver read requests")
	command.Flags().Int("webserver.write-timeout", 60, "timeout for webserver write requests")
	command.Flags().Int("webserver.idle-timeout", 60, "timeout for webserver idle requests")
	command.Flags().Int("webserver.request-timeout", 60, "seconds allowed for total request processing before the request context is cancelled (0 disables the limit)")
	command.Flags().Int("webserver.shutdown-timeout", 15, "seconds a graceful shutdown waits for in-flight requests to drain before forcing close (0 keeps this server's safe 15s default)")
	command.Flags().Int("webserver.max-header-bytes", 1<<20, "maximum size in bytes of request headers the server will read (bounds memory per request; 0 keeps this server's safe 1 MiB default)")
	command.Flags().Int64("webserver.max-body-bytes", 10<<20, "maximum size in bytes of each request body the server will read (bounds memory per request; 0 disables the limit)")
	command.Flags().Bool("webserver.pprof-enabled", false, "expose /debug/pprof/* profiling endpoints on the public mux (unauthenticated; enable only in trusted environments)")
	command.Flags().StringSlice("webserver.cors-allowed-origins", []string{"https://*", "http://*"}, "origins permitted by CORS (supports '*' wildcards); the default allows any origin — restrict to your front-end origins for production")
	command.Flags().String("webserver.tls-cert-file", "", "path to a PEM-encoded TLS certificate; set with tls-key-file to terminate TLS in-process (leave unset when a reverse proxy terminates TLS)")
	command.Flags().String("webserver.tls-key-file", "", "path to the PEM-encoded TLS private key; set with tls-cert-file to terminate TLS in-process (leave unset when a reverse proxy terminates TLS)")

	// Database options
	command.Flags().String("database.uri", "", "PostgreSQL connection URI (e.g. postgres://user:pass@host:5432/db)")
	command.Flags().Int("database.ping-timeout", 15, "timeout in seconds for database ping operations")
	command.Flags().Int("database.max-conns", 0, "maximum size of the connection pool (0 uses the pgx default: max(4, NumCPU))")
	command.Flags().Int("database.min-conns", 0, "number of idle connections kept warm in the pool (0 uses the pgx default: 0)")
	command.Flags().Int("database.max-conn-lifetime", 0, "seconds after which a connection is closed and replaced (0 uses the pgx default: 3600)")
	command.Flags().Int("database.max-conn-idle-time", 0, "seconds after which an idle connection is closed (0 uses the pgx default: 1800)")
	command.Flags().Int("database.health-check-period", 0, "seconds between pool health checks of idle connections (0 uses the pgx default: 60)")
	command.Flags().Int("database.close-timeout", 10, "seconds to wait for the connection pool to close on shutdown before abandoning it (0 waits indefinitely)")

	// Auth options
	command.Flags().String("auth.jwt-secret", "", "shared HMAC secret for signing/verifying JWT bearer tokens (protected endpoints reject all requests when empty)")
	command.Flags().String("auth.jwt-issuer", "", "required issuer (iss) claim on validated JWTs; not checked when empty")
	command.Flags().String("auth.jwt-audience", "", "required audience (aud) claim on validated JWTs; not checked when empty")
	command.Flags().Int("auth.jwt-ttl", 3600, "lifetime in seconds of generated JWTs")

	// Logging options
	command.Flags().String("logging.level", "info", "minimum log level: debug, info, warn, or error (--verbose forces debug)")
	command.Flags().String("logging.format", "", "log encoding: json or console (defaults to console in development, json otherwise)")

	// Observability options
	command.Flags().Bool("observability.enabled", false, "export traces and metrics via OTLP/gRPC (disabled by default; when off the OpenTelemetry providers stay no-op and nothing is exported)")
	command.Flags().String("observability.endpoint", "", "OTLP/gRPC collector address as host:port without scheme (empty uses OTEL_EXPORTER_OTLP_ENDPOINT, then the SDK default localhost:4317)")
	command.Flags().Bool("observability.insecure", false, "disable TLS on the OTLP/gRPC connection (use only for local collectors or trusted networks)")
	command.Flags().Float64("observability.sample-ratio", 1.0, "head-based trace sampling probability in [0.0, 1.0], wrapped in a parent-based sampler (1.0 samples every trace, 0.0 none)")

	// Lifecycle options
	command.Flags().Int("shutdown-delay", 0, "seconds to wait after a shutdown signal before draining (e.g. for Kubernetes endpoint sync)")
}

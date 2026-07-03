package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rfizzle/go-starter/internal/app"
	"github.com/rfizzle/go-starter/internal/cli"
	"go.uber.org/zap"

	"github.com/rfizzle/go-starter/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cryptCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the go starter application.",
	Long: `Run the go starter application. 

Example Command:
go-starter serve --config /etc/crypt/config.json`,
	PersistentPreRun: persistentPreRunHook(nil),
	PreRun: func(cmd *cobra.Command, args []string) {
		// Webserver options
		cobra.CheckErr(viper.BindPFlag("webserver.address", cmd.Flags().Lookup("webserver-address")))
		cobra.CheckErr(viper.BindPFlag("webserver.port", cmd.Flags().Lookup("webserver-port")))
		cobra.CheckErr(viper.BindPFlag("webserver.read-timeout", cmd.Flags().Lookup("webserver-read-timeout")))
		cobra.CheckErr(viper.BindPFlag("webserver.write-timeout", cmd.Flags().Lookup("webserver-write-timeout")))
		cobra.CheckErr(viper.BindPFlag("webserver.idle-timeout", cmd.Flags().Lookup("webserver-idle-timeout")))

		// Database options
		cobra.CheckErr(viper.BindPFlag("database.uri", cmd.Flags().Lookup("database-uri")))
		cobra.CheckErr(viper.BindPFlag("database.ping-timeout", cmd.Flags().Lookup("database-ping-timeout")))

		// Auth options
		cobra.CheckErr(viper.BindPFlag("auth.jwt-secret", cmd.Flags().Lookup("auth-jwt-secret")))
		cobra.CheckErr(viper.BindPFlag("auth.jwt-issuer", cmd.Flags().Lookup("auth-jwt-issuer")))
		cobra.CheckErr(viper.BindPFlag("auth.jwt-audience", cmd.Flags().Lookup("auth-jwt-audience")))
		cobra.CheckErr(viper.BindPFlag("auth.jwt-ttl", cmd.Flags().Lookup("auth-jwt-ttl")))

		// Lifecycle options
		cobra.CheckErr(viper.BindPFlag("shutdown-delay", cmd.Flags().Lookup("shutdown-delay")))
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Setup config with CLI/ENV variables
		cfg := &config.Application{
			Build: config.Build{
				Branch:   BuildBranch,
				Date:     BuildDate,
				Env:      BuildEnv,
				Revision: BuildRevision,
				Version:  BuildVersion,
			},
			Webserver: config.Webserver{
				Address:      viper.GetString("webserver.address"),
				Port:         viper.GetInt("webserver.port"),
				ReadTimeout:  viper.GetInt("webserver.read-timeout"),
				WriteTimeout: viper.GetInt("webserver.write-timeout"),
				IdleTimeout:  viper.GetInt("webserver.idle-timeout"),
			},
			Database: config.Database{
				URI:         viper.GetString("database.uri"),
				PingTimeout: viper.GetInt("database.ping-timeout"),
			},
			Auth: config.Auth{
				JWTSecret:   viper.GetString("auth.jwt-secret"),
				JWTIssuer:   viper.GetString("auth.jwt-issuer"),
				JWTAudience: viper.GetString("auth.jwt-audience"),
				JWTTTL:      viper.GetInt("auth.jwt-ttl"),
			},
			ShutdownDelay: viper.GetInt("shutdown-delay"),
		}

		// Configure the application
		application, err := app.New(cfg)
		if err != nil {
			logger.Fatal("error initializing application", zap.Error(err))
		}

		// Set up the application
		application.SetLogger(logger)

		// Root context cancelled on SIGINT/SIGTERM. WithGracefulShutdown layers
		// the readiness flip and configurable drain delay on top before the
		// app's components are asked to stop.
		signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		runCtx, cancel := cli.WithGracefulShutdown(signalCtx, application, time.Duration(cfg.ShutdownDelay)*time.Second)
		defer cancel()

		// Run the application; blocks until shutdown or a component fails.
		if err := application.Run(runCtx); err != nil {
			logger.Error("error running application", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Webserver options
	serveCmd.Flags().String("webserver-address", "0.0.0.0", "interface to which the webserver will bind")
	serveCmd.Flags().Int("webserver-port", 8080, "port on which the webserver will listen")
	serveCmd.Flags().Int("webserver-read-timeout", 60, "timeout for webserver read requests")
	serveCmd.Flags().Int("webserver-write-timeout", 60, "timeout for webserver write requests")
	serveCmd.Flags().Int("webserver-idle-timeout", 60, "timeout for webserver idle requests")

	// Database options
	serveCmd.Flags().String("database-uri", "", "PostgreSQL connection URI (e.g. postgres://user:pass@host:5432/db)")
	serveCmd.Flags().Int("database-ping-timeout", 15, "timeout in seconds for database ping operations")

	// Auth options
	serveCmd.Flags().String("auth-jwt-secret", "", "shared HMAC secret for signing/verifying JWT bearer tokens (protected endpoints reject all requests when empty)")
	serveCmd.Flags().String("auth-jwt-issuer", "", "required issuer (iss) claim on validated JWTs; not checked when empty")
	serveCmd.Flags().String("auth-jwt-audience", "", "required audience (aud) claim on validated JWTs; not checked when empty")
	serveCmd.Flags().Int("auth-jwt-ttl", 3600, "lifetime in seconds of generated JWTs")

	// Lifecycle options
	serveCmd.Flags().Int("shutdown-delay", 0, "seconds to wait after a shutdown signal before draining (e.g. for Kubernetes endpoint sync)")
}

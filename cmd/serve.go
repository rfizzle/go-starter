package cmd

import (
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
	PreRun: func(cmd *cobra.Command, args []string) {
		// Webserver options
		cobra.CheckErr(viper.BindPFlag("webserver.address", cmd.Flags().Lookup("webserver-address")))
		cobra.CheckErr(viper.BindPFlag("webserver.port", cmd.Flags().Lookup("webserver-port")))
		cobra.CheckErr(viper.BindPFlag("webserver.read-timeout", cmd.Flags().Lookup("webserver-read-timeout")))
		cobra.CheckErr(viper.BindPFlag("webserver.write-timeout", cmd.Flags().Lookup("webserver-write-timeout")))
		cobra.CheckErr(viper.BindPFlag("webserver.idle-timeout", cmd.Flags().Lookup("webserver-idle-timeout")))
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
		}

		// Configure the application
		application, err := app.New(cfg)
		if err != nil {
			logger.Fatal("error initializing application", zap.Error(err))
		}

		// Support graceful shutdown
		cli.GracefulExit(application)

		// Start the application
		err = application.Start()
		if err != nil {
			logger.Error("error starting application", zap.Error(err))
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
}

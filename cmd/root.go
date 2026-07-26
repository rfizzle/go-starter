// Package cmd defines the Cobra command tree and the process-level concerns
// around it: configuration file and environment binding through Viper, logger
// construction, and the build metadata linked in at release time.
//
// Commands own argument validation and process exit behavior only. Each one
// assembles its dependencies through internal/app or a focused composition
// root of its own and then hands off; no business logic lives here.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ApplicationName string
	BuildBranch     string
	BuildDate       string
	BuildEnv        string
	BuildRevision   string
	BuildVersion    string

	cfgFile string
	logger  = zap.NewNop()
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "go-starter",
	Short:   "A starter golang project",
	Version: BuildVersion,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		loggerConfig := zap.NewProductionConfig()
		var err error
		if verbose {
			loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		} else {
			loggerConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		}
		logger, err = loggerConfig.Build()
		if err != nil || logger == nil {
			panic(fmt.Errorf("failed to initialize logger: %w", err))
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "log debug level messages")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	replacer := strings.NewReplacer(".", "_", "-", "_")
	viper.SetEnvPrefix("GOSTARTER")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv() // read in environment variables that match
}

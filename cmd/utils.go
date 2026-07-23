package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rfizzle/go-starter/internal/platform/config"
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

func persistentPreRunHook(fn func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if cmd.HasParent() {
			cmd.Parent().PersistentPreRun(cmd.Parent(), args)
		}
		if fn != nil {
			fn(cmd, args)
		}
	}
}

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

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rfizzle/go-starter/pkg/config"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// resolveConfig unmarshals every viper layer (flag > env > file > default) into
// a config.Application, stamps the build metadata that lives in ldflags rather
// than config, and applies the --verbose debug shortcut. It is the shared
// front-half of every long-running subcommand (serve, seed) so config handling
// stays identical across them.
func resolveConfig() (*config.Application, error) {
	cfg := &config.Application{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
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

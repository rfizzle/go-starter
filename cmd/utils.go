package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
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

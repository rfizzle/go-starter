// Command initialize replaces the Go Starter template identity.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/rfizzle/go-starter/internal/templateinit"
)

func main() {
	var cfg templateinit.Config

	flags := flag.NewFlagSet("initialize", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&cfg.ModulePath, "module", "", "Go module path (required)")
	flags.StringVar(&cfg.ApplicationName, "app-name", "", "human-readable application name (required)")
	flags.StringVar(&cfg.BinaryName, "binary", "", "binary and release name (required)")
	flags.StringVar(&cfg.EnvPrefix, "env-prefix", "", "environment-variable prefix (required)")
	flags.StringVar(&cfg.ImageName, "image", "", "untagged OCI image name (required)")
	flags.StringVar(&cfg.Description, "description", "", "short application description")
	flags.StringVar(&cfg.Root, "root", ".", "template repository root")
	flags.Usage = func() {
		fmt.Fprintln(flags.Output(), "Initialize a clean Go Starter clone exactly once.")
		fmt.Fprintln(flags.Output())
		fmt.Fprintln(flags.Output(), "Usage:")
		fmt.Fprintln(flags.Output(), "  go run ./tools/initialize [flags]")
		fmt.Fprintln(flags.Output())
		flags.PrintDefaults()
	}

	if err := flags.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "initialize: unexpected arguments: %v\n", flags.Args())
		flags.Usage()
		os.Exit(2)
	}

	if err := templateinit.Initialize(context.Background(), cfg); err != nil {
		fmt.Fprintf(os.Stderr, "initialize: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Template initialized successfully.")
	fmt.Println("Review the changes, then run: make generate-check && make check && make build")
}

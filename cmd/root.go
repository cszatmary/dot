package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/cszatmary/dot/client"
	"github.com/cszatmary/dot/internal/log"
	"github.com/spf13/cobra"
)

// Set by goreleaser when release build is created.
var version string

// Execute runs the dot CLI.
func Execute() {
	var c container
	rootCmd := newRootCommand(&c)
	if err := rootCmd.Execute(); err != nil {
		if c.opts.verbose {
			fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
		os.Exit(1)
	}
}

// container stores all the dependencies that can be used by commands.
type container struct {
	logger    *log.Logger
	dotClient *client.Client
	opts      struct {
		verbose bool
	}
}

func newRootCommand(c *container) *cobra.Command {
	// Set version if built from source
	if version == "" {
		version = "source"
		if info, available := debug.ReadBuildInfo(); available {
			version = info.Main.Version
		}
	}
	rootCmd := &cobra.Command{
		Use:     "dot",
		Version: version,
		Short:   "dot is a CLI for managing dotfiles.",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		// cobra prints errors returned from RunE by default. Disable that since we handle errors ourselves.
		SilenceErrors: true,
		// cobra prints command usage by default if RunE returns an error.
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			c.logger = log.New(os.Stderr)
			c.logger.SetDebug(c.opts.verbose)
			dotClient, err := client.New(client.WithDebugger(c.logger))
			if err != nil {
				return fmt.Errorf("failed to setup dot: %w", err)
			}
			c.dotClient = dotClient
			return nil
		},
	}
	rootCmd.AddCommand(
		newApplyCommand(c),
		newCompletionsCommand(),
		newSetupCommand(c),
	)
	rootCmd.PersistentFlags().BoolVarP(&c.opts.verbose, "verbose", "v", false, "enable verbose output")
	return rootCmd
}

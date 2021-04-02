package cmd

import (
	"github.com/TouchBistro/goutils/fatal"
	"github.com/cszatmary/dot/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Set by goreleaser when release build is created
var version string

type rootOptions struct {
	verbose bool
}

var (
	rootOpts  rootOptions
	logger    = logrus.StandardLogger()
	dotClient *client.Client
)

var rootCmd = &cobra.Command{
	Use:     "dot",
	Version: version,
	Short:   "dot is a CLI for managing dotfiles.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		fatal.ShowStackTraces(rootOpts.verbose)
		if rootOpts.verbose {
			logger.SetLevel(logrus.DebugLevel)
		}
		logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: true,
		})
		var err error
		dotClient, err = client.New(client.WithDebugger(logger))
		if err != nil {
			fatal.ExitErr(err, "Failed to initialize dot")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootOpts.verbose, "verbose", "v", false, "enable verbose output")
}

// Execute runs the dot CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fatal.ExitErr(err, "Failed executing command.")
	}
}

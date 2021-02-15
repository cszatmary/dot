package cmd

import (
	"github.com/TouchBistro/goutils/color"
	"github.com/TouchBistro/goutils/fatal"
	"github.com/cszatma/dot/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Set by goreleaser when release build is created
var version string

type rootOptions struct {
	verbose bool
}

var (
	rootOpts rootOptions
	logger   = logrus.StandardLogger()
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

		// ACTION: this doesn't belong here
		if cmd.Name() != "setup" && !config.IsSetup() {
			fatal.Exit(color.Red("Error: dot has not been setup. Please run `dot setup`."))
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&rootOpts.verbose, "verbose", "v", false, "enable verbose output")
	cobra.OnInitialize(func() {
		err := config.Init()
		if err != nil {
			fatal.ExitErr(err, "Failed to read config file")
		}
	})
}

// Execute runs the dot CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fatal.ExitErr(err, "Failed executing command.")
	}
}

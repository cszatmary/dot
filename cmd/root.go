package cmd

import (
	"github.com/TouchBistro/goutils/color"
	"github.com/TouchBistro/goutils/fatal"
	"github.com/cszatma/dot/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var (
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:     "dot",
	Version: version,
	Short:   "dot is a CLI for managing dotfiles",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() != "setup" && !config.IsSetup() {
			fatal.Exit(color.Red("Error: dot has not been setup. Please run `dot setup`."))
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	cobra.OnInitialize(func() {
		var logLevel log.Level
		if verbose {
			logLevel = log.DebugLevel
		} else {
			logLevel = log.InfoLevel
			fatal.ShowStackTraces = false
		}

		log.SetLevel(logLevel)
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp: true,
		})

		err := config.Init()
		if err != nil {
			fatal.ExitErr(err, "Failed to read config file")
		}
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fatal.ExitErr(err, "Failed executing command.")
	}
}

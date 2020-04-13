package cmd

import (
	"fmt"

	"github.com/TouchBistro/goutils/color"
	"github.com/TouchBistro/goutils/fatal"
	"github.com/cszatma/dot/config"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [DOTFILES...]",
	Args:  cobra.ArbitraryArgs,
	Short: "Apply dotfile changes",
	Run: func(cmd *cobra.Command, args []string) {
		dotfiles := config.Config().Dotfiles

		// Get list of dotfiles to apply
		var dotfileNames []string
		if len(args) == 0 {
			dotfileNames = make([]string, 0, len(dotfiles))
			for name := range dotfiles {
				dotfileNames = append(dotfileNames, name)
			}
		} else {
			dotfileNames = make([]string, 0, len(args))
			for _, arg := range args {
				if _, ok := dotfiles[arg]; !ok {
					fatal.Exitf(color.Red("Invalid dotfile %s\n"), arg)
				}

				dotfileNames = append(dotfileNames, arg)
			}
		}

		// Apply changes
		for _, name := range dotfileNames {
			fmt.Printf(color.Cyan("Applying %s\n"), name)
			// Save last known hash in lockfile so we can warn the user that they might overwrite changed shit
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}

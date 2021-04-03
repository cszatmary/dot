package cmd

import (
	"github.com/TouchBistro/goutils/fatal"
	"github.com/spf13/cobra"
)

type applyOptions struct {
	force bool
}

var applyOpts applyOptions

var applyCmd = &cobra.Command{
	Use:   "apply [DOTFILES...]",
	Args:  cobra.ArbitraryArgs,
	Short: "Apply dotfile changes",
	Run: func(cmd *cobra.Command, args []string) {
		if !dotClient.IsSetup() {
			fatal.Exit("dot has not been setup. Please run `dot setup`.")
		}

		logger.Printf("Applying changes to dotfiles")
		err := dotClient.Apply(applyOpts.force, args...)
		if err != nil {
			fatal.ExitErr(err, "Failed to apply changes to dotfiles")
		}
		logger.Printf("Successfully applied changes to dotfiles")
	},
}

func init() {
	applyCmd.Flags().BoolVarP(&applyOpts.force, "force", "f", false, "Overwrite dotfile if it was manually modified")
	rootCmd.AddCommand(applyCmd)
}

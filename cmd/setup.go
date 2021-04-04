package cmd

import (
	"github.com/TouchBistro/goutils/fatal"
	"github.com/spf13/cobra"
)

type setupOptions struct {
	registryPath string
	force        bool
}

var setupOpts setupOptions

var setupCmd = &cobra.Command{
	Use:   "setup",
	Args:  cobra.NoArgs,
	Short: "Setup dot to manage your dotfiles",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Printf("Setting up dot...")
		err := dotClient.Setup(setupOpts.registryPath, setupOpts.force)
		if err != nil {
			fatal.ExitErr(err, "Failed to setup dot")
		}
		logger.Printf("Successfully setup dot")
	},
}

func init() {
	setupCmd.Flags().StringVarP(&setupOpts.registryPath, "registry", "r", "~/.dotfiles", "path to directory where dotfile sources are located")
	setupCmd.Flags().BoolVarP(&setupOpts.force, "force", "f", false, "Re-setup dot with a new dotfiles source")
	rootCmd.AddCommand(setupCmd)
}

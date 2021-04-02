package cmd

import (
	"os"
	"path/filepath"

	"github.com/TouchBistro/goutils/fatal"
	"github.com/spf13/cobra"
)

type setupOptions struct {
	dotfilesPath string
	force        bool
}

var setupOpts setupOptions

var setupCmd = &cobra.Command{
	Use:   "setup",
	Args:  cobra.NoArgs,
	Short: "Setup dot to manage your dotfiles",
	Run: func(cmd *cobra.Command, args []string) {
		if setupOpts.dotfilesPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fatal.ExitErr(err, "Failed to find user home directory")
			}
			setupOpts.dotfilesPath = filepath.Join(homeDir, ".dotfiles")
		}

		logger.Info("Setting up dot...")
		err := dotClient.Setup(setupOpts.dotfilesPath, setupOpts.force)
		if err != nil {
			fatal.ExitErr(err, "Failed to setup dot")
		}
		logger.Info("Successfully setup dot")
	},
}

func init() {
	setupCmd.Flags().StringVarP(&setupOpts.dotfilesPath, "dotfiles-path", "d", "", "path to directory where dotfile sources are located")
	setupCmd.Flags().BoolVarP(&setupOpts.force, "force", "f", false, "Re-setup dot with a new dotfiles source")
	rootCmd.AddCommand(setupCmd)
}

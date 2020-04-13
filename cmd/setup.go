package cmd

import (
	"os"
	"path/filepath"

	"github.com/TouchBistro/goutils/color"
	"github.com/TouchBistro/goutils/fatal"
	"github.com/cszatma/dot/config"
	log "github.com/sirupsen/logrus"
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
		if config.IsSetup() && !setupOpts.force {
			log.Infoln("dot is already setup. If you wish to set it up again, use the --force flag.")
			return
		}

		log.Infoln(color.Cyan("Setting up dot..."))
		err := config.Setup(setupOpts.dotfilesPath)
		if err != nil {
			fatal.ExitErr(err, "Failed to setup dot")
		}

		log.Infoln(color.Green("Successfully setup dot"))
	},
}

func init() {
	defaultPath := filepath.Join(os.Getenv("HOME"), ".dotfiles")
	setupCmd.Flags().StringVarP(&setupOpts.dotfilesPath, "dotfiles-path", "d", defaultPath, "path to directory where dotfile sources are located")
	setupCmd.Flags().BoolVarP(&setupOpts.force, "force", "f", false, "Setup even if already setup")
	rootCmd.AddCommand(setupCmd)
}

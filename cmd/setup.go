package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TouchBistro/goutils/color"
	"github.com/TouchBistro/goutils/fatal"
	"github.com/TouchBistro/goutils/file"
	"github.com/cszatma/dot/config"
	"github.com/spf13/cobra"
)

var (
	dotfilesPath string
	force        bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Args:  cobra.NoArgs,
	Short: "Setup dot to manage your dotfiles",
	Run: func(cmd *cobra.Command, args []string) {
		if config.IsSetup() && !force {
			fmt.Println("dot is already setup. If you wish to set it up again, use the --force flag.")
			return
		}

		err := config.Setup(dotfilesPath)
		if err != nil {
			fatal.ExitErr(err, "Failed to setup dot")
		}

		for name, dotfile := range config.Config().Dotfiles {
			fmt.Printf(color.Cyan("Creating backup of %s\n"), name)
			backupPath := dotfile.Dest + ".bak"
			err = file.CopyFile(dotfile.Dest, backupPath)
			if err != nil {
				fatal.ExitErrf(err, "Failed to create backup of %s at %s", name, backupPath)
			}

			fmt.Printf(color.Green("Created backup of %s at %s\n"), name, backupPath)
		}

		fmt.Println(color.Green("Successfully setup dot"))
	},
}

func init() {
	defaultPath := filepath.Join(os.Getenv("HOME"), ".dotfiles")
	setupCmd.Flags().StringVarP(&dotfilesPath, "dotfiles-path", "d", defaultPath, "path to directory where dotfile sources are located")
	setupCmd.Flags().BoolVarP(&force, "force", "f", false, "Setup even if already setup")
	rootCmd.AddCommand(setupCmd)
}

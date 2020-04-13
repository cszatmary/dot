package cmd

import (
	"strings"

	"github.com/TouchBistro/goutils/color"
	"github.com/TouchBistro/goutils/fatal"
	"github.com/cszatma/dot/config"
	log "github.com/sirupsen/logrus"
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

		log.Infof("Applying changes to the following dotfiles: %s", strings.Join(dotfileNames, ", "))
		err := config.Apply(dotfileNames, applyOpts.force)
		if err != nil {
			fatal.ExitErr(err, "Failed to apply changes to dotfiles")
		}
		log.Infoln(color.Green("Successfully applied changes to dotfiles"))
	},
}

func init() {
	applyCmd.Flags().BoolVarP(&applyOpts.force, "force", "f", false, "Overwrite dotfile if it was manually modified")
	rootCmd.AddCommand(applyCmd)
}

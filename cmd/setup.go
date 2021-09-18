package cmd

import (
	"github.com/spf13/cobra"
)

func newSetupCommand(c *container) *cobra.Command {
	var setupOpts struct {
		registryPath string
		force        bool
	}
	setupCmd := &cobra.Command{
		Use:   "setup",
		Args:  cobra.NoArgs,
		Short: "Setup dot to manage your dotfiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			c.logger.Printf("Setting up dot...")
			err := c.dotClient.Setup(setupOpts.registryPath, setupOpts.force)
			if err != nil {
				return err
			}
			c.logger.Printf("Successfully setup dot")
			return nil
		},
	}
	setupCmd.Flags().StringVarP(&setupOpts.registryPath, "registry", "r", "~/.dotfiles", "path to directory where dotfile sources are located")
	setupCmd.Flags().BoolVarP(&setupOpts.force, "force", "f", false, "Re-setup dot with a new dotfiles source")
	return setupCmd
}

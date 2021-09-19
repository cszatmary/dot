package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newApplyCommand(c *container) *cobra.Command {
	var applyOpts struct {
		force bool
	}
	applyCmd := &cobra.Command{
		Use:   "apply [DOTFILES...]",
		Args:  cobra.ArbitraryArgs,
		Short: "Apply dotfile changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !c.dotClient.IsSetup() {
				return fmt.Errorf("dot has not been setup, run `dot setup` to set it up")
			}
			c.logger.Printf("Applying changes to dotfiles")
			err := c.dotClient.Apply(applyOpts.force, args...)
			if err != nil {
				return err
			}
			c.logger.Printf("Successfully applied changes to dotfiles")
			return nil
		},
	}
	applyCmd.Flags().BoolVarP(&applyOpts.force, "force", "f", false, "Overwrite dotfile if it was manually modified")
	return applyCmd
}

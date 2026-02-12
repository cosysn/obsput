package cmd

import "github.com/spf13/cobra"

func NewOBSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obs",
		Short: "Manage OBS configurations",
	}
	return cmd
}

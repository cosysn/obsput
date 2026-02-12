package cmd

import (
	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <version>",
		Short: "Delete a version",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("delete command")
		},
	}
	cmd.Flags().String("name", "", "OBS name to delete from")
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewDeleteCommand())
}

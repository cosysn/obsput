package cmd

import (
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List uploaded versions",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("list command")
		},
	}
	cmd.Flags().StringP("output", "o", "table", "Output format (table/json)")
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewListCommand())
}

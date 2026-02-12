package cmd

import (
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Version:", version)
			cmd.Println("Commit:", commit)
			cmd.Println("Date:", date)
		},
	}
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewVersionCommand())
}

package cmd

import (
	"github.com/spf13/cobra"
)

func NewDownloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <version>",
		Short: "Show download commands for a version",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("download command")
		},
	}
	cmd.Flags().StringP("output", "o", "", "Output file path")
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewDownloadCommand())
}

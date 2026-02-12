package cmd

import (
	"github.com/spf13/cobra"
)

func NewUploadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload binary to OBS",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("upload command")
		},
	}
	cmd.Flags().String("prefix", "", "Path prefix for upload")
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewUploadCommand())
}

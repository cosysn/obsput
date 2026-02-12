package cmd

import (
	"github.com/spf13/cobra"
)

var version = "dev"
var commit = "unknown"
var date = "unknown"

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "obsput",
		Short:   "Upload binaries to Huawei Cloud OBS",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(NewOBSCommand())
	return cmd
}

func Execute() {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

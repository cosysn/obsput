package cmd

import (
	"obsput/pkg/config"

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
	cmd.AddCommand(NewUploadCommand())
	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewDeleteCommand())
	cmd.AddCommand(NewDownloadCommand())
	return cmd
}

func Execute() {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func getConfigPath() string {
	path, _ := config.GetConfigPath()
	return path
}

func getConfigDir() string {
	dir, _ := config.GetConfigDir()
	return dir
}

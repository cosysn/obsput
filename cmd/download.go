package cmd

import (
	"fmt"

	"obsput/pkg/config"
	obsclient "obsput/pkg/obs"

	"github.com/spf13/cobra"
)

func NewDownloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <version>",
		Short: "Show download commands for a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			outputPath, _ := cmd.Flags().GetString("output")

			// Load config
			cfg, err := config.Load(getConfigPath())
			if err != nil {
				return fmt.Errorf("load config failed: %v", err)
			}

			for name, obsCfg := range cfg.Configs {
				cmd.Printf("[%s]\n", name)

				client := obsclient.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)

				// Find the version
				versions, err := client.ListVersions("")
				if err != nil {
					cmd.Printf("Error: %v\n", err)
					continue
				}

				for _, v := range versions {
					if v.Version == version {
						cmd.Printf("Version: %s\n", v.Version)
						cmd.Printf("URL: %s\n", v.URL)
						cmd.Println()
						cmd.Println("Commands:")
						cmd.Printf("  curl -O %s\n", v.URL)
						cmd.Printf("  wget %s\n", v.URL)
						if outputPath != "" {
							cmd.Printf("  curl -o %s %s\n", outputPath, v.URL)
						}
					}
				}
			}

			return nil
		},
	}
	cmd.Flags().StringP("output", "o", "", "Output file path")
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewDownloadCommand())
}

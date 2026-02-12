package cmd

import (
	"fmt"
	obsclient "obsput/pkg/obs"

	"obsput/pkg/config"
	"obsput/pkg/output"

	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List uploaded versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			outputFormat, _ := cmd.Flags().GetString("output")

			// Load config
			cfg, err := config.Load(getConfigPath())
			if err != nil {
				return fmt.Errorf("load config failed: %v", err)
			}

			formatter := output.NewFormatter()

			for name, obsCfg := range cfg.Configs {
				cmd.Printf("[%s]\n", name)

				client := obsclient.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)
				versions, err := client.ListVersions("")
				if err != nil {
					cmd.Printf("Error: %v\n", err)
					continue
				}

				items := make([]output.VersionItem, 0, len(versions))
				for _, v := range versions {
					items = append(items, output.VersionItem{
						Version: v.Version,
						Size:    v.Size,
						Date:    v.Date,
						Commit:  v.Commit,
						URL:     v.URL,
					})
				}

				if outputFormat == "json" {
					formatter.PrintJSON(items)
				} else {
					formatter.PrintVersionTable(items)
				}
			}

			return nil
		},
	}
	cmd.Flags().StringP("output", "o", "table", "Output format (table/json)")
	return cmd
}

func init() {
	root := NewRootCommand()
	root.AddCommand(NewListCommand())
}

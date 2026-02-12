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
			profile, _ := cmd.Flags().GetString("profile")

			// Load config
			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("No OBS configurations configured\n\nConfig file: %s\n\nAdd OBS:\n  obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"", getConfigPath())
			}

			// Determine which configs to list
			var configsToUse map[string]*config.OBS
			if profile != "" {
				obsCfg := cfg.GetOBS(profile)
				if obsCfg == nil {
					return fmt.Errorf("profile '%s' not found in config\n\nRun: obsput obs list", profile)
				}
				configsToUse = map[string]*config.OBS{
					profile: obsCfg,
				}
			} else {
				configsToUse = cfg.Configs
			}

			formatter := output.NewFormatter()

			for name, obsCfg := range configsToUse {
				cmd.Println()
				cmd.Printf("  [%s]\n", name)
				cmd.Println()

				client := obsclient.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)
				versions, err := client.ListVersions("")
				if err != nil {
					cmd.Printf("  Error: %v\n", err)
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
	cmd.Flags().StringP("profile", "p", "", "OBS profile name to use (default: all profiles)")
	return cmd
}

func init() {}

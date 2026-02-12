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
			profile, _ := cmd.Flags().GetString("profile")

			// Load config
			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("No OBS configurations configured\n\nConfig file: %s\n\nAdd OBS:\n  obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"", getConfigPath())
			}

			// Determine which configs to use
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

			for name, obsCfg := range configsToUse {
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
	cmd.Flags().StringP("profile", "p", "", "OBS profile name to use (default: all profiles)")
	return cmd
}

func init() {}

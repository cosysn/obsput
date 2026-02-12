package cmd

import (
	"fmt"

	"obsput/pkg/config"
	obsclient "obsput/pkg/obs"

	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <version>",
		Short: "Delete a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			targetName, _ := cmd.Flags().GetString("name")

			// Load config
			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("No OBS configurations configured\n\nConfig file: %s\n\nAdd OBS:\n  obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"", getConfigPath())
			}

			for name, obsCfg := range cfg.Configs {
				// If name specified, only delete from that OBS
				if targetName != "" && name != targetName {
					continue
				}

				cmd.Printf("[%s] Deleting %s...\n", name, version)

				client := obsclient.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)
				result := client.DeleteVersion(version)

				if result.Success {
					cmd.Printf("[%s] Deleted: %s\n", name, version)
				} else {
					cmd.Printf("[%s] Failed: %s\n", name, result.Error)
				}
			}

			return nil
		},
	}
	cmd.Flags().String("name", "", "OBS name to delete from")
	return cmd
}

func init() {}

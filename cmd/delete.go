package cmd

import (
	"fmt"

	"obsput/pkg/config"
	obsclient "obsput/pkg/obs"
	"obsput/pkg/styled"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <version>",
		Short: "Delete a version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			profile, _ := cmd.Flags().GetString("profile")

			// Load config
			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("No OBS configurations configured\n\nConfig file: %s\n\nAdd OBS:\n  obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"", getConfigPath())
			}

			// Create styled output
			out := styled.NewOutput()

			out.Divider()
			out.Section("Delete Version")
			out.KeyValue("Version", version)
			out.Divider()

			deletedCount := 0
			failedCount := 0

			for name, obsCfg := range cfg.Configs {
				// If profile specified, only delete from that OBS
				if profile != "" && name != profile {
					continue
				}

				out.Subsection("[" + name + "]")

				client := obsclient.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)
				result := client.DeleteVersion(version)

				t := table.NewWriter()
				t.SetOutputMirror(cmd.OutOrStdout())
				t.AppendHeader(table.Row{"Field", "Value"})
				if result.Success {
					t.AppendRow(table.Row{"Status", "Deleted"})
					out.SuccessMsg(fmt.Sprintf("%s: deleted", obsCfg.Bucket))
					deletedCount++
				} else {
					t.AppendRow(table.Row{"Status", "Failed"})
					t.AppendRow(table.Row{"Error", result.Error})
					out.ErrorMsg(fmt.Sprintf("%s: %s", obsCfg.Bucket, result.Error))
					failedCount++
				}
				t.Render()
				out.Divider()
			}

			out.Section("Summary")
			out.Summary(deletedCount, failedCount)

			return nil
		},
	}
	cmd.Flags().StringP("profile", "p", "", "OBS profile name to use (default: all profiles)")
	return cmd
}

func init() {}

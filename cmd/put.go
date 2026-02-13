package cmd

import (
	"fmt"
	"os"
	"time"

	"obsput/pkg/config"
	"obsput/pkg/obs"
	"obsput/pkg/output"
	versionpkg "obsput/pkg/version"
	"obsput/pkg/progress"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewPutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put <file>",
		Short: "Put binary to OBS",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			prefix, _ := cmd.Flags().GetString("prefix")
			profile, _ := cmd.Flags().GetString("profile")

			// Check file exists
			fileInfo, err := os.Stat(filePath)
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", filePath)
			}

			// Generate version
			gen := versionpkg.NewGenerator()
			ver := gen.Generate()

			// Load config
			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("No OBS configurations configured\n\nConfig file: %s\n\nAdd OBS:\n  obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"", getConfigPath())
			}

			// Determine which configs to put to
			var configsToUse map[string]*config.OBS
			if profile != "" {
				// Use specific profile
				obsCfg := cfg.GetOBS(profile)
				if obsCfg == nil {
					return fmt.Errorf("profile '%s' not found in config\n\nRun: obsput obs list", profile)
				}
				configsToUse = map[string]*config.OBS{
					profile: obsCfg,
				}
			} else {
				// Put to all configs
				configsToUse = cfg.Configs
			}

			// Print header
			cmd.Println()
			cmd.Println("  File:", filePath)
			cmd.Println("  Version:", ver)
			cmd.Println()

			// Create progress bar
			pb := progress.New(fileInfo.Size())
			formatter := output.NewFormatter()

			// Put to selected OBS configs
			successCount := 0
			failCount := 0

			for name, obsCfg := range configsToUse {
				cmd.Printf("  [%s]\n", name)

				client := obs.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)

				startTime := time.Now()

				result, err := client.UploadFile(filePath, ver, prefix, func(bytes int64) {
					pb.SetTotal(fileInfo.Size())
					pb.Increment(bytes - pb.Current())
					pb.Render()
				})

				// Clear progress bar line
				cmd.Println()

				if err != nil {
					cmd.Printf("    Failed: %v\n", err)
					failCount++
					continue
				}

				if result.Success {
					// Print result in a nice table format
					t := table.NewWriter()
					t.SetOutputMirror(cmd.OutOrStdout())
					t.AppendHeader(table.Row{"Field", "Value"})
					t.AppendRow(table.Row{"URL", result.URL})
					t.AppendRow(table.Row{"Signed URL (24h)", result.SignedURL})
					t.AppendRow(table.Row{"MD5", result.MD5})
					t.AppendRow(table.Row{"Size", formatter.FormatSize(result.Size)})
					if result.Size > 0 {
						elapsed := time.Since(startTime)
						speed := float64(result.Size) / elapsed.Seconds()
						t.AppendRow(table.Row{"Speed", formatter.FormatSize(int64(speed)) + "/s"})
					}
					t.Render()

					// Show download commands
					cmd.Println()
					cmd.Println("  Download:")
					cmd.Printf("    wget %s -O <filename>\n", result.SignedURL)
					cmd.Printf("    curl -L %s -o <filename>\n", result.SignedURL)
					successCount++
				} else {
					cmd.Printf("    Failed: %s\n", result.Error)
					failCount++
				}
				cmd.Println()
			}

			cmd.Printf("%d completed, %d failed\n", successCount, failCount)

			return nil
		},
	}
	cmd.Flags().String("prefix", "", "Path prefix for put")
	cmd.Flags().StringP("profile", "p", "", "OBS profile name to use (default: all profiles)")
	return cmd
}

func init() {}

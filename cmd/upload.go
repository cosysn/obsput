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

	"github.com/spf13/cobra"
)

func NewUploadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload binary to OBS",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			prefix, _ := cmd.Flags().GetString("prefix")

			// Check file exists
			fileInfo, err := os.Stat(filePath)
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", filePath)
			}

			// Generate version
			gen := versionpkg.NewGenerator()
			ver := gen.Generate()

			// Load config
			cfg, err := config.Load(getConfigPath())
			if err != nil {
				return fmt.Errorf("load config failed: %v", err)
			}

			// Upload to all OBS configs
			cmd.Println("Uploading:", filePath)
			cmd.Println("Version:", ver)
			cmd.Println()

			// Create progress bar
			pb := progress.New(fileInfo.Size())
			formatter := output.NewFormatter()

			// Upload to all OBS configs
			successCount := 0
			failCount := 0

			for name, obsCfg := range cfg.Configs {
				cmd.Printf("[%s]\n", name)

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
					cmd.Printf("  Failed: %v\n", err)
					failCount++
					continue
				}

				if result.Success {
					cmd.Printf("  Uploaded: %s\n", result.URL)
					cmd.Printf("  MD5: %s\n", result.MD5)
					cmd.Printf("  Size: %s\n", formatter.FormatSize(result.Size))
					if result.Size > 0 {
						elapsed := time.Since(startTime)
						speed := float64(result.Size) / elapsed.Seconds()
						cmd.Printf("  Speed: %s/s\n", formatter.FormatSize(int64(speed)))
					}
					successCount++
				} else {
					cmd.Printf("  Failed: %s\n", result.Error)
					failCount++
				}
				cmd.Println()
			}

			cmd.Printf("%d completed, %d failed\n", successCount, failCount)

			return nil
		},
	}
	cmd.Flags().String("prefix", "", "Path prefix for upload")
	return cmd
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return home + "/.obsput.yaml"
}

func init() {}

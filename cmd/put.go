package cmd

import (
	"fmt"
	"os"
	"time"

	"obsput/pkg/config"
	"obsput/pkg/obs"
	"obsput/pkg/output"
	"obsput/pkg/styled"
	versionpkg "obsput/pkg/version"
	"obsput/pkg/progress"

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

			// Create styled output
			out := styled.NewOutput()
			formatter := output.NewFormatter()

			// Print header
			out.Divider()
			out.Section("Upload")
			out.KeyValue("File", filePath)
			out.KeyValue("Version", ver)
			out.Divider()

			// Create progress bar
			pb := progress.New(fileInfo.Size())

			// Put to selected OBS configs
			successCount := 0
			failCount := 0

			for name, obsCfg := range configsToUse {
				out.Subsection("[" + name + "]")

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
					out.ErrorMsg(fmt.Sprintf("Upload failed: %v", err))
					failCount++
					continue
				}

				if result.Success {
					// Extract filename from key
					filename := client.ExtractFilenameFromKey(result.Key)
					// Print result in a styled box
					content := map[string]string{
						"URL":        result.URL,
						"Size":       formatter.FormatSize(result.Size),
						"MD5":        result.MD5,
						"Clean URL":  obs.CleanURL(result.SignedURL),
					}
					if result.Size > 0 {
						elapsed := time.Since(startTime)
						speed := float64(result.Size) / elapsed.Seconds()
						content["Speed"] = formatter.FormatSize(int64(speed)) + "/s"
					}
					out.PrintBox("Upload Result", content)

					out.Println(styled.Header, "Download Commands:")
					cleanURL := obs.CleanURL(result.SignedURL)
					out.Printf(styled.Muted, "  curl -k -o %s %s\n", filename, cleanURL)
					out.Printf(styled.Muted, "  wget --no-check-certificate -O %s %s\n", filename, cleanURL)
					out.Spacer()
					out.SuccessMsg(fmt.Sprintf("Uploaded to %s", obsCfg.Bucket))
					successCount++
				} else {
					out.ErrorMsg(result.Error)
					failCount++
				}
				cmd.Println()
			}

			out.Section("Summary")
			out.Summary(successCount, failCount)

			return nil
		},
	}
	cmd.Flags().String("prefix", "", "Path prefix for put")
	cmd.Flags().StringP("profile", "p", "", "OBS profile name to use (default: all profiles)")
	return cmd
}

func init() {}

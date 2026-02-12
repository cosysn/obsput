package cmd

import (
	"fmt"
	"os"

	"obsput/pkg/config"
	"obsput/pkg/obs"
	versionpkg "obsput/pkg/version"

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
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
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

			for name, obsCfg := range cfg.Configs {
				client := obs.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)

				result, err := client.UploadFile(filePath, ver, prefix, nil)
				if err != nil {
					cmd.Printf("[%s] Failed: %v\n", name, err)
					continue
				}

				if result.Success {
					cmd.Printf("[%s] Uploaded: %s\n", name, result.URL)
					cmd.Printf("  MD5: %s\n", result.MD5)
				} else {
					cmd.Printf("[%s] Failed: %s\n", name, result.Error)
				}
			}

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

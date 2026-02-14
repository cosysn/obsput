package cmd

import (
	"fmt"

	"obsput/pkg/config"
	obsclient "obsput/pkg/obs"
	"obsput/pkg/styled"

	"github.com/spf13/cobra"
)

func NewDownloadCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <version>",
		Short: "Show download commands for a version",
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
				return fmt.Errorf("no OBS configurations configured.\n\nRun: obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"")
			}

			// Determine which configs to use
			var configsToUse map[string]*config.OBS
			if profile != "" {
				obsCfg := cfg.GetOBS(profile)
				if obsCfg == nil {
					return fmt.Errorf("profile '%s' not found in config.\n\nRun: obsput obs list", profile)
				}
				configsToUse = map[string]*config.OBS{
					profile: obsCfg,
				}
			} else {
				configsToUse = cfg.Configs
			}

			// Create styled output
			out := styled.NewOutput()

			out.Divider()
			out.Section("Download")
			out.KeyValue("Version", version)
			out.Divider()

			found := false
			for name, obsCfg := range configsToUse {
				out.Subsection("[" + name + "]")

				client := obsclient.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)

				// Find the version
				versions, err := client.ListVersions("")
				if err != nil {
					out.ErrorMsg(fmt.Sprintf("Failed to list versions: %v", err))
					continue
				}

				for _, v := range versions {
					if v.Version == version {
						found = true
						cleanURL := obsclient.CleanURL(v.URL)
						out.KeyValue("Version", v.Version)
						out.KeyValue("URL", cleanURL)
						out.KeyValue("Size", v.Size)
						out.Divider()
						out.Println(styled.Header, "Download Commands:")
						out.Printf(styled.Muted, "  curl -k -o <filename> %s\n", cleanURL)
						out.Printf(styled.Muted, "  wget --no-check-certificate %s\n", cleanURL)
					}
				}
			}

			if !found {
				out.WarningMsg(fmt.Sprintf("Version %s not found", version))
			}

			return nil
		},
	}
	cmd.Flags().StringP("profile", "p", "", "OBS profile name to use (default: all profiles)")
	return cmd
}

func init() {}

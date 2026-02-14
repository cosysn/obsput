package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"obsput/pkg/config"
	obsclient "obsput/pkg/obs"
	"obsput/pkg/styled"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <version|--before>",
		Short: "Delete a version or versions before a date",
		Long: `Delete a specific version or all versions before a specified date.

Examples:
  # Delete specific version
  obsput delete v1.0.0-abc123-20260214-153045-1

  # Delete all versions before 2026-01-01
  obsput delete --before 2026-01-01

  # Delete all versions older than 7 days
  obsput delete --before 7d

  # Delete all versions older than 24 hours
  obsput delete --before 24h`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			before, _ := cmd.Flags().GetString("before")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			// Load config
			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("no OBS configurations configured.\n\nRun: obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"")
			}

			// Parse before date
			var beforeTime time.Time
			if before != "" {
				beforeTime, err = parseBeforeDate(before)
				if err != nil {
					return fmt.Errorf("invalid --before format: %v\nUse format: YYYY-MM-DD or Nd (e.g., 7d, 24h)", err)
				}
			}

			// Create styled output
			out := styled.NewOutput()

			out.Divider()
			if before != "" {
				out.Section(fmt.Sprintf("Delete Before %s", beforeTime.Format("2006-01-02")))
				if dryRun {
					out.WarningMsg("DRY RUN - No versions will be deleted")
					out.Spacer()
				}
			} else if len(args) > 0 {
				out.Section(fmt.Sprintf("Delete %s", args[0]))
			} else {
				return fmt.Errorf("specify a version or use --before to delete versions by date")
			}
			out.Divider()

			totalDeleted := 0
			totalFailed := 0

			for name, obsCfg := range cfg.Configs {
				// If profile specified, only delete from that OBS
				if profile != "" && name != profile {
					continue
				}

				out.Subsection("[" + name + "]")

				client := obsclient.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)

				// List all versions
				versions, err := client.ListVersions("")
				if err != nil {
					out.ErrorMsg(fmt.Sprintf("Failed to list versions: %v", err))
					continue
				}

				// Filter versions to delete
				var toDelete []string
				if before != "" {
					for _, v := range versions {
						versionDate, err := parseVersionDate(v.Version)
						if err == nil && versionDate.Before(beforeTime) {
							toDelete = append(toDelete, v.Version)
						}
					}
				} else if len(args) > 0 {
					// Delete specific version (prefix match)
					targetVersion := args[0]
					for _, v := range versions {
						if strings.HasPrefix(v.Version, targetVersion) {
							toDelete = append(toDelete, v.Version)
						}
					}
				}

				if len(toDelete) == 0 {
					out.Println(styled.Muted, "  No versions to delete")
					continue
				}

				// Show versions to delete
				out.Println(styled.Info, fmt.Sprintf("  Found %d version(s) to delete:", len(toDelete)))
				for _, v := range toDelete {
					out.Println(styled.Muted, fmt.Sprintf("    - %s", v))
				}
				out.Spacer()

				if dryRun {
					out.WarningMsg("DRY RUN - Skipping deletion")
					continue
				}

				// Delete versions
				deleted := 0
				failed := 0
				for _, v := range toDelete {
					result := client.DeleteVersion(v)
					if result.Success {
						deleted++
						out.SuccessMsg(fmt.Sprintf("Deleted: %s", v))
					} else {
						failed++
						out.ErrorMsg(fmt.Sprintf("Failed: %s (%s)", v, result.Error))
					}
				}

				// Summary for this OBS
				t := table.NewWriter()
				t.SetOutputMirror(cmd.OutOrStdout())
				t.AppendHeader(table.Row{"Metric", "Count"})
				t.AppendRow(table.Row{"Deleted", deleted})
				t.AppendRow(table.Row{"Failed", failed})
				t.Render()

				totalDeleted += deleted
				totalFailed += failed
				out.Divider()
			}

			// Final summary
			if before != "" && !dryRun {
				out.Section("Summary")
				out.Summary(totalDeleted, totalFailed)
			}

			return nil
		},
	}
	cmd.Flags().StringP("profile", "p", "", "OBS profile name to use (default: all profiles)")
	cmd.Flags().String("before", "", "Delete versions before this date (YYYY-MM-DD or Nd, Nh)")
	cmd.Flags().Bool("dry-run", false, "Show what would be deleted without actually deleting")
	return cmd
}

// parseBeforeDate parses date string like "2026-01-01" or "7d" or "24h"
func parseBeforeDate(s string) (time.Time, error) {
	// Try date format YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	// Try relative format Nd (days) or Nh (hours)
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err == nil {
			return time.Now().AddDate(0, 0, -days), nil
		}
	}
	if strings.HasSuffix(s, "h") {
		hours, err := strconv.Atoi(strings.TrimSuffix(s, "h"))
		if err == nil {
			return time.Now().Add(-time.Duration(hours) * time.Hour), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid format: %s", s)
}

// parseVersionDate extracts date from version string
// Format: v1.0.0-commit-YYYYMMDD-HHMMSS-counter
func parseVersionDate(version string) (time.Time, error) {
	parts := strings.Split(version, "-")
	if len(parts) < 4 {
		return time.Time{}, fmt.Errorf("invalid version format")
	}

	// parts[2] is date like "20260214"
	dateStr := parts[2]
	if len(dateStr) != 8 {
		return time.Time{}, fmt.Errorf("invalid date in version")
	}

	return time.Parse("20060102", dateStr)
}

func init() {}

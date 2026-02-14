package cmd

import (
	"fmt"
	"sync"

	"obsput/pkg/config"
	"obsput/pkg/obs"

	"github.com/spf13/cobra"
)

func NewOBSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obs",
		Short: "Manage OBS configurations",
	}
	cmd.AddCommand(NewOBSAddCommand())
	cmd.AddCommand(NewOBSListCommand())
	cmd.AddCommand(NewOBSGetCommand())
	cmd.AddCommand(NewOBSRemoveCommand())
	cmd.AddCommand(NewOBSInitCommand())
	cmd.AddCommand(NewOBSMakeBucketCommand())
	return cmd
}

func NewOBSAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add OBS configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			endpoint, _ := cmd.Flags().GetString("endpoint")
			bucket, _ := cmd.Flags().GetString("bucket")
			ak, _ := cmd.Flags().GetString("ak")
			sk, _ := cmd.Flags().GetString("sk")

			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v", err)
			}

			cfg.AddOBS(name, endpoint, bucket, ak, sk)

			if err := cfg.Save(getConfigPath()); err != nil {
				return fmt.Errorf("save config failed: %v", err)
			}

			cmd.Printf("Added OBS config: %s\n", name)
			return nil
		},
	}
	cmd.Flags().String("name", "", "OBS name")
	cmd.Flags().String("endpoint", "", "OBS endpoint")
	cmd.Flags().String("bucket", "", "OBS bucket")
	cmd.Flags().String("ak", "", "Access Key")
	cmd.Flags().String("sk", "", "Secret Key")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("bucket")
	cmd.MarkFlagRequired("ak")
	cmd.MarkFlagRequired("sk")
	return cmd
}

func NewOBSListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OBS configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v\nRun: obsput obs add --name prod --endpoint \"xxx\" --bucket \"xxx\" --ak \"xxx\" --sk \"xxx\"", err)
			}

			cmd.Println("NAME\tENDPOINT\tBUCKET\tSTATUS")
			for _, obs := range cfg.ListOBS() {
				cmd.Printf("%s\t%s\t%s\tactive\n", obs.Name, obs.Endpoint, obs.Bucket)
			}
			return nil
		},
	}
	return cmd
}

func NewOBSGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Get OBS configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v", err)
			}

			obs := cfg.GetOBS(name)
			if obs == nil {
				return fmt.Errorf("OBS config not found: %s", name)
			}

			cmd.Printf("Name: %s\n", obs.Name)
			cmd.Printf("Endpoint: %s\n", obs.Endpoint)
			cmd.Printf("Bucket: %s\n", obs.Bucket)
			cmd.Printf("AK: %s\n", maskAK(obs.AK))
			cmd.Printf("SK: %s\n", maskSK(obs.SK))
			return nil
		},
	}
	return cmd
}

func NewOBSRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove OBS configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v", err)
			}

			if !cfg.OBSExists(name) {
				return fmt.Errorf("OBS config not found: %s", name)
			}

			cfg.RemoveOBS(name)

			if err := cfg.Save(getConfigPath()); err != nil {
				return fmt.Errorf("save config failed: %v", err)
			}

			cmd.Printf("Removed OBS config: %s\n", name)
			return nil
		},
	}
	return cmd
}

func NewOBSInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize OBS configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getConfigPath()

			cfg := config.NewConfig()
			if err := cfg.Ensure(path); err != nil {
				return fmt.Errorf("init config failed: %v", err)
			}

			cmd.Printf("Initialized config: %s\n", path)
			cmd.Println("\nAdd OBS configuration:")
			cmd.Println("  obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"")
			return nil
		},
	}
	return cmd
}

func NewOBSMakeBucketCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mb [name]",
		Short: "Create bucket for OBS",
		Long: `Create bucket for specified OBS profile.
If no profile is specified, creates bucket for all configured OBS.
Returns error if bucket already exists.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadOrInit()
			if err != nil {
				return fmt.Errorf("load config failed: %v", err)
			}

			if len(cfg.Configs) == 0 {
				return fmt.Errorf("no OBS configurations configured\n\nRun: obsput obs add --name prod --endpoint \"obs.xxx.com\" --bucket \"bucket\" --ak \"xxx\" --sk \"xxx\"")
			}

			// Determine which configs to use
			var configsToUse map[string]*config.OBS
			if len(args) > 0 && args[0] != "" {
				// Use specific profile
				name := args[0]
				obsCfg := cfg.GetOBS(name)
				if obsCfg == nil {
					return fmt.Errorf("OBS config not found: %s\n\nRun: obsput obs list", name)
				}
				configsToUse = map[string]*config.OBS{
					name: obsCfg,
				}
			} else {
				// All configs
				configsToUse = cfg.Configs
			}

			cmd.Println()
			cmd.Println("Creating buckets...")
			cmd.Println()

			// Create buckets concurrently
			var mu sync.Mutex
			var wg sync.WaitGroup
			results := make([]*obs.BucketResult, 0, len(configsToUse))

			for name, obsCfg := range configsToUse {
				wg.Add(1)
				go func(name string, obsCfg *config.OBS) {
					defer wg.Done()

					client := obs.NewClient(obsCfg.Endpoint, obsCfg.Bucket, obsCfg.AK, obsCfg.SK)

					err := client.CreateBucket()
					result := &obs.BucketResult{
						OBSName: name,
						Bucket:  obsCfg.Bucket,
						Success: err == nil,
					}
					if err != nil {
						result.Error = err.Error()
					}

					mu.Lock()
					results = append(results, result)
					mu.Unlock()
				}(name, obsCfg)
			}

			wg.Wait()

			// Print results
			successCount := 0
			failCount := 0
			for _, r := range results {
				if r.Success {
					cmd.Printf("[%s] Created: %s\n", r.OBSName, r.Bucket)
					successCount++
				} else {
					cmd.Printf("[%s] Failed: %s (%s)\n", r.OBSName, r.Bucket, r.Error)
					failCount++
				}
			}

			cmd.Println()
			cmd.Printf("%d created, %d failed\n", successCount, failCount)
			return nil
		},
	}
	cmd.Flags().StringP("profile", "p", "", "OBS profile name to use (default: all profiles)")
	return cmd
}

func maskAK(ak string) string {
	if len(ak) <= 4 {
		return "****"
	}
	return ak[:len(ak)-4] + "****"
}

func maskSK(sk string) string {
	if len(sk) <= 4 {
		return "****"
	}
	return sk[:len(sk)-4] + "****"
}

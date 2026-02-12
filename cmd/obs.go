package cmd

import (
	"fmt"

	"obsput/pkg/config"

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
			cfg, err := config.Load(getConfigPath())
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

			cfg, err := config.Load(getConfigPath())
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

			cfg, err := config.Load(getConfigPath())
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

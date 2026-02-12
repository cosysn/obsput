package cmd

import (
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
	return cmd
}

func NewOBSAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add OBS configuration",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	cmd.Flags().String("name", "", "OBS name")
	cmd.Flags().String("endpoint", "", "OBS endpoint")
	cmd.Flags().String("bucket", "", "OBS bucket")
	cmd.Flags().String("ak", "", "Access Key")
	cmd.Flags().String("sk", "", "Secret Key")
	return cmd
}

func NewOBSListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OBS configurations",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	return cmd
}

func NewOBSGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get OBS configuration",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	return cmd
}

func NewOBSRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove OBS configuration",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	return cmd
}

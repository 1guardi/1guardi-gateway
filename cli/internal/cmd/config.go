package cmd

import (
	cliconfig "github.com/chaitanyabankanhal/ai-gateway/cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  "Read and write ~/.config/aigw/config.yaml.\n\nKeys: endpoint, tenant",
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Print a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		val, err := cliconfig.Get(args[0])
		if err != nil {
			return err
		}
		printf(cmd, "%s\n", val)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Write a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cliconfig.Set(args[0], args[1]); err != nil {
			return err
		}
		printf(cmd, "Set %s = %s\n", args[0], args[1])
		return nil
	},
}

func init() {
	configCmd.AddCommand(configGetCmd, configSetCmd)
}

package cmd

import (
	"github.com/chaitanyabankanhal/ai-gateway/cli/internal/auth"
	"github.com/spf13/cobra"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Print the stored JWT (for curl/scripts)",
	Long:  "Print the raw session JWT to stdout, e.g. for use with curl:\n  curl -H \"Authorization: Bearer $(aigw token)\" ...",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		creds, err := auth.Load()
		if err != nil {
			return err
		}
		printf(cmd, "%s\n", creds.Token)
		return nil
	},
}

package cmd

import (
	"github.com/chaitanyabankanhal/ai-gateway/cli/internal/auth"
	cliconfig "github.com/chaitanyabankanhal/ai-gateway/cli/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear the stored session",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := auth.Clear(); err != nil {
			return err
		}
		// Best-effort: drop the cached proxy key so no plaintext gateway
		// key lingers in the keyring after logout.
		if cfg, err := cliconfig.Load(); err == nil && cfg.Tenant != "" {
			_ = auth.ClearProxyKey(cfg.Tenant)
		}
		printf(cmd, "Logged out.\n")
		return nil
	},
}

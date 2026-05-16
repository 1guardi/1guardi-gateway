package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/chaitanyabankanhal/ai-gateway/cli/internal/client"
	cliconfig "github.com/chaitanyabankanhal/ai-gateway/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	keysTenant  string
	keysName    string
	keysAgentID uint
	keysUserID  uint
	keysID      string
	keysJSON    bool
)

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Manage tenant API keys",
}

var keysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys for a tenant",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		tenant, err := resolveTenant()
		if err != nil {
			return err
		}
		c, _, err := authedClient()
		if err != nil {
			return err
		}
		keys, err := c.ListKeys(cmd.Context(), tenant)
		if err != nil {
			return err
		}

		if keysJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(keys)
		}
		if len(keys) == 0 {
			printf(cmd, "No API keys for tenant %s.\n", tenant)
			return nil
		}
		tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		_, _ = tw.Write([]byte("ID\tNAME\tKEY\tACTIVE\tLAST USED\n"))
		for _, k := range keys {
			lastUsed := "never"
			if k.LastUsedAt != nil {
				lastUsed = k.LastUsedAt.Local().Format("2006-01-02 15:04")
			}
			masked := fmt.Sprintf("%s_…%s", k.Prefix, k.Suffix)
			_, _ = tw.Write([]byte(formatRow(k.ID, k.Name, masked, fmt.Sprintf("%t", k.IsActive), lastUsed)))
		}
		return tw.Flush()
	},
}

var keysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an API key (plaintext shown once)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if keysName == "" {
			return fmt.Errorf("--name is required")
		}
		tenant, err := resolveTenant()
		if err != nil {
			return err
		}
		c, _, err := authedClient()
		if err != nil {
			return err
		}

		req := client.CreateKeyRequest{Name: keysName}
		if cmd.Flags().Changed("agent") {
			req.AgentID = &keysAgentID
		}
		if cmd.Flags().Changed("user") {
			req.UserID = &keysUserID
		}

		created, err := c.CreateKey(cmd.Context(), tenant, req)
		if err != nil {
			return err
		}
		printf(cmd, "Created API key %q (id %d).\n\n", created.Name, created.ID)
		printf(cmd, "  %s\n\n", created.Key)
		printf(cmd, "This is the only time the key is shown — store it now.\n")
		return nil
	},
}

var keysRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke (deactivate) an API key",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if keysID == "" {
			return fmt.Errorf("--id is required")
		}
		tenant, err := resolveTenant()
		if err != nil {
			return err
		}
		c, _, err := authedClient()
		if err != nil {
			return err
		}
		if err := c.RevokeKey(cmd.Context(), tenant, keysID); err != nil {
			return err
		}
		printf(cmd, "Revoked API key %s.\n", keysID)
		return nil
	},
}

// resolveTenant returns the --tenant flag value, falling back to the
// configured default tenant.
func resolveTenant() (string, error) {
	if keysTenant != "" {
		return keysTenant, nil
	}
	cfg, err := cliconfig.Load()
	if err != nil {
		return "", err
	}
	if cfg.Tenant == "" {
		return "", fmt.Errorf("no tenant specified — pass --tenant or set one with `aigw config set tenant <id>`")
	}
	return cfg.Tenant, nil
}

func init() {
	for _, c := range []*cobra.Command{keysListCmd, keysCreateCmd, keysRevokeCmd} {
		c.Flags().StringVar(&keysTenant, "tenant", "", "tenant ID (defaults to configured tenant)")
	}
	keysListCmd.Flags().BoolVar(&keysJSON, "json", false, "output as JSON")
	keysCreateCmd.Flags().StringVar(&keysName, "name", "", "key name (required)")
	keysCreateCmd.Flags().UintVar(&keysAgentID, "agent", 0, "scope the key to an agent ID")
	keysCreateCmd.Flags().UintVar(&keysUserID, "user", 0, "scope the key to a user ID")
	keysRevokeCmd.Flags().StringVar(&keysID, "id", "", "API key ID to revoke (required)")

	keysCmd.AddCommand(keysListCmd, keysCreateCmd, keysRevokeCmd)
}

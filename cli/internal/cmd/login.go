package cmd

import (
	"fmt"

	"github.com/chaitanyabankanhal/ai-gateway/cli/internal/auth"
	"github.com/chaitanyabankanhal/ai-gateway/cli/internal/client"
	cliconfig "github.com/chaitanyabankanhal/ai-gateway/cli/internal/config"
	"github.com/spf13/cobra"
)

var loginProvider string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in via browser-based SSO",
	Long: "Log in to the AI Gateway via browser-based SSO.\n\n" +
		"Opens your browser to the configured identity provider, captures the\n" +
		"resulting session token on a local loopback listener, and stores it in\n" +
		"your OS keyring.",
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := cliconfig.Load()
		if err != nil {
			return err
		}
		ctx := cmd.Context()

		provider := loginProvider
		if provider == "" {
			provider, err = pickProvider(cmd, cfg.Endpoint)
			if err != nil {
				return err
			}
		}

		token, err := auth.BrowserLogin(ctx, cfg.Endpoint, provider)
		if err != nil {
			return err
		}

		id, err := auth.DecodeIdentity(token)
		if err != nil {
			return err
		}
		if err := auth.Save(auth.Credentials{
			Token:    token,
			Endpoint: cfg.Endpoint,
			Expiry:   id.Expiry,
		}); err != nil {
			return err
		}

		printf(cmd, "Logged in as %s <%s>\n", id.Name, id.Email)
		if !id.Expiry.IsZero() {
			printf(cmd, "Session valid until %s\n", id.Expiry.Local().Format("2006-01-02 15:04 MST"))
		}
		return nil
	},
}

// pickProvider resolves which OIDC provider to use when --provider is omitted:
// it auto-selects a sole enabled provider, else lists the options.
func pickProvider(cmd *cobra.Command, endpoint string) (string, error) {
	providers, err := client.FetchProviders(cmd.Context(), endpoint)
	if err != nil {
		return "", err
	}
	switch len(providers) {
	case 0:
		return "", fmt.Errorf("no SSO providers are enabled on the gateway")
	case 1:
		return providers[0].Name, nil
	default:
		msg := "multiple SSO providers enabled — choose one with --provider:\n"
		for _, p := range providers {
			msg += fmt.Sprintf("  %s (%s)\n", p.Name, p.Label)
		}
		return "", fmt.Errorf("%s", msg)
	}
}

func init() {
	loginCmd.Flags().StringVar(&loginProvider, "provider", "", "OIDC provider name (e.g. google, microsoft)")
}

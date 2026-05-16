package cmd

import (
	"time"

	"github.com/chaitanyabankanhal/ai-gateway/cli/internal/auth"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the current logged-in identity",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		creds, err := auth.Load()
		if err != nil {
			return err
		}
		id, err := auth.DecodeIdentity(creds.Token)
		if err != nil {
			return err
		}

		printf(cmd, "Name:     %s\n", id.Name)
		printf(cmd, "Email:    %s\n", id.Email)
		printf(cmd, "User ID:  %d\n", id.UserID)
		printf(cmd, "Endpoint: %s\n", creds.Endpoint)
		if id.IsSuperAdmin {
			printf(cmd, "Role:     super admin\n")
		}
		if !id.Expiry.IsZero() {
			status := "valid"
			if time.Now().After(id.Expiry) {
				status = "EXPIRED — run `aigw login`"
			}
			printf(cmd, "Session:  %s (%s)\n", id.Expiry.Local().Format("2006-01-02 15:04 MST"), status)
		}
		return nil
	},
}

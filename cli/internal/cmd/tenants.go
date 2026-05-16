package cmd

import (
	"encoding/json"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var tenantsJSON bool

var tenantsCmd = &cobra.Command{
	Use:   "tenants",
	Short: "Manage tenants",
}

var tenantsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tenants visible to you",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		c, _, err := authedClient()
		if err != nil {
			return err
		}
		tenants, err := c.ListTenants(cmd.Context())
		if err != nil {
			return err
		}

		if tenantsJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(tenants)
		}
		if len(tenants) == 0 {
			printf(cmd, "No tenants.\n")
			return nil
		}
		tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		_, _ = tw.Write([]byte("ID\tNAME\tDESCRIPTION\n"))
		for _, t := range tenants {
			_, _ = tw.Write([]byte(formatRow(t.ID, t.Name, t.Description)))
		}
		return tw.Flush()
	},
}

func init() {
	tenantsListCmd.Flags().BoolVar(&tenantsJSON, "json", false, "output as JSON")
	tenantsCmd.AddCommand(tenantsListCmd)
}

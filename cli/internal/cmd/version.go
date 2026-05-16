package cmd

import "github.com/spf13/cobra"

// Build metadata, injected via -ldflags at build time.
var (
	version = "dev"
	commit  = "none"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the aigw version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		printf(cmd, "aigw %s (%s)\n", version, commit)
		return nil
	},
}

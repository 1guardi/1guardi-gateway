// Package cmd implements the aigw CLI commands.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "aigw",
	Short:         "AI Gateway CLI — SSO login and API key management",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command, printing any error to stderr.
func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		rootCmd.PrintErrln("Error:", err)
	}
	return err
}

func init() {
	rootCmd.AddCommand(
		loginCmd,
		logoutCmd,
		whoamiCmd,
		tokenCmd,
		tenantsCmd,
		keysCmd,
		envCmd,
		runCmd,
		configCmd,
		versionCmd,
	)
}

// printf writes to the command's stdout.
func printf(cmd *cobra.Command, format string, a ...any) {
	fmt.Fprintf(cmd.OutOrStdout(), format, a...)
}

// Command aigw is the AI Gateway CLI for SSO login and API key management.
package main

import (
	"os"

	"github.com/chaitanyabankanhal/ai-gateway/cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

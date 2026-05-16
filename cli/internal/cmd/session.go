package cmd

import (
	"fmt"
	"os"

	"github.com/chaitanyabankanhal/ai-gateway/cli/internal/auth"
	"github.com/chaitanyabankanhal/ai-gateway/cli/internal/client"
	cliconfig "github.com/chaitanyabankanhal/ai-gateway/cli/internal/config"
)

// authedClient loads the stored session and returns an admin API client.
// It errors with a clear message when the user is not logged in.
func authedClient() (*client.Client, *auth.Credentials, error) {
	creds, err := auth.Load()
	if err != nil {
		return nil, nil, err
	}
	if creds.Expired() {
		fmt.Fprintln(os.Stderr, "Warning: session has expired — run `aigw login`.")
	}
	endpoint := creds.Endpoint
	if endpoint == "" {
		cfg, err := cliconfig.Load()
		if err != nil {
			return nil, nil, err
		}
		endpoint = cfg.Endpoint
	}
	return client.New(endpoint, creds.Token), creds, nil
}

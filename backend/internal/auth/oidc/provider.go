package oidc

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	goidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Provider wraps a configured OIDC provider with its OAuth2 config + ID-token verifier.
type Provider struct {
	Name     string // "google" | "microsoft"
	Label    string // human-readable, e.g. "Google"
	OAuth2   *oauth2.Config
	Verifier *goidc.IDTokenVerifier
}

// Registry is a thread-safe lookup of enabled providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]*Provider
}

// NewRegistry constructs a registry, performing OIDC discovery for every enabled provider.
// Discovery happens once at startup so callback handlers stay on the hot path with cached metadata.
func NewRegistry(ctx context.Context, cfg config.OIDCConfig) (*Registry, error) {
	r := &Registry{providers: make(map[string]*Provider)}

	if cfg.Google.Enabled() {
		p, err := buildGoogle(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("oidc: google: %w", err)
		}
		r.providers["google"] = p
	}
	if cfg.Microsoft.Enabled() {
		p, err := buildMicrosoft(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("oidc: microsoft: %w", err)
		}
		r.providers["microsoft"] = p
	}
	return r, nil
}

func (r *Registry) Get(name string) (*Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

// Enabled returns the sorted list of enabled provider names.
func (r *Registry) Enabled() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for n := range r.providers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func redirectURL(base, provider string) string {
	return base + "/api/v1/auth/oidc/" + provider + "/callback"
}

func buildGoogle(ctx context.Context, cfg config.OIDCConfig) (*Provider, error) {
	prov, err := goidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("discovery: %w", err)
	}
	return &Provider{
		Name:  "google",
		Label: "Google",
		OAuth2: &oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			Endpoint:     prov.Endpoint(),
			RedirectURL:  redirectURL(cfg.RedirectBaseURL, "google"),
			Scopes:       []string{goidc.ScopeOpenID, "email", "profile"},
		},
		Verifier: prov.Verifier(&goidc.Config{ClientID: cfg.Google.ClientID}),
	}, nil
}

func buildMicrosoft(ctx context.Context, cfg config.OIDCConfig) (*Provider, error) {
	issuer := "https://login.microsoftonline.com/" + cfg.MicrosoftTenantID + "/v2.0"
	prov, err := goidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("discovery: %w", err)
	}
	verifierCfg := &goidc.Config{ClientID: cfg.Microsoft.ClientID}
	// The "common" / "organizations" / "consumers" Microsoft endpoints issue tokens with
	// tenant-specific issuers, so the issuer in the ID token won't match the discovery issuer.
	// Skip the issuer check when running multi-tenant; verify clientID + signature + expiry instead.
	if cfg.MicrosoftTenantID == "common" || cfg.MicrosoftTenantID == "organizations" || cfg.MicrosoftTenantID == "consumers" {
		verifierCfg.SkipIssuerCheck = true
	}
	return &Provider{
		Name:  "microsoft",
		Label: "Microsoft",
		OAuth2: &oauth2.Config{
			ClientID:     cfg.Microsoft.ClientID,
			ClientSecret: cfg.Microsoft.ClientSecret,
			Endpoint:     prov.Endpoint(),
			RedirectURL:  redirectURL(cfg.RedirectBaseURL, "microsoft"),
			Scopes:       []string{goidc.ScopeOpenID, "email", "profile"},
		},
		Verifier: prov.Verifier(verifierCfg),
	}, nil
}

// Claims is the subset of standard OIDC claims we extract from verified ID tokens.
type Claims struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	// Microsoft sometimes returns email as "preferred_username" when "email" scope is absent.
	PreferredUsername string `json:"preferred_username"`
}

// ResolvedEmail prefers verified email, falling back to preferred_username.
func (c Claims) ResolvedEmail() string {
	if c.Email != "" {
		return c.Email
	}
	return c.PreferredUsername
}

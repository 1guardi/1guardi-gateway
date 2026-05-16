// Package client is a thin HTTP client for the AI Gateway admin API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrUnauthorized is returned when the admin API rejects the session token.
var ErrUnauthorized = errors.New("session expired or invalid — run `aigw login`")

// Client talks to the admin API with a bearer JWT.
type Client struct {
	endpoint string
	token    string
	http     *http.Client
}

// New builds a client for the given admin API base URL and JWT.
func New(endpoint, token string) *Client {
	return &Client{
		endpoint: strings.TrimRight(endpoint, "/"),
		token:    token,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// Tenant is a gateway tenant as returned by the admin API.
type Tenant struct {
	ID          uint   `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
}

// APIKey is a tenant API key. The plaintext key is never returned on listing.
type APIKey struct {
	ID         uint       `json:"ID"`
	Name       string     `json:"Name"`
	Prefix     string     `json:"Prefix"`
	Suffix     string     `json:"Suffix"`
	TenantID   uint       `json:"TenantID"`
	AgentID    *uint      `json:"AgentID"`
	UserID     *uint      `json:"UserID"`
	IsActive   bool       `json:"IsActive"`
	LastUsedAt *time.Time `json:"LastUsedAt"`
	CreatedAt  time.Time  `json:"CreatedAt"`
}

// CreateKeyRequest is the body for creating an API key.
type CreateKeyRequest struct {
	Name    string `json:"name"`
	AgentID *uint  `json:"agent_id,omitempty"`
	UserID  *uint  `json:"user_id,omitempty"`
}

// CreatedKey is the create-key response, including the one-time plaintext key.
type CreatedKey struct {
	APIKey
	Key string `json:"key"`
}

// Provider is an enabled OIDC login provider (public, unauthenticated).
type Provider struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// FetchProviders lists the gateway's enabled OIDC providers. No auth required.
func FetchProviders(ctx context.Context, endpoint string) ([]Provider, error) {
	url := strings.TrimRight(endpoint, "/") + "/api/v1/auth/oidc/providers"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("reach gateway at %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list providers: %s", resp.Status)
	}
	var out []Provider
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode providers: %w", err)
	}
	return out, nil
}

// ListTenants returns the tenants visible to the caller.
func (c *Client) ListTenants(ctx context.Context) ([]Tenant, error) {
	var out []Tenant
	return out, c.do(ctx, http.MethodGet, "/api/v1/tenants", nil, &out)
}

// ListKeys returns the API keys for a tenant.
func (c *Client) ListKeys(ctx context.Context, tenantID string) ([]APIKey, error) {
	var out []APIKey
	return out, c.do(ctx, http.MethodGet, "/api/v1/tenants/"+tenantID+"/keys", nil, &out)
}

// CreateKey creates a new API key and returns the one-time plaintext key.
func (c *Client) CreateKey(ctx context.Context, tenantID string, req CreateKeyRequest) (*CreatedKey, error) {
	var out CreatedKey
	if err := c.do(ctx, http.MethodPost, "/api/v1/tenants/"+tenantID+"/keys", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RevokeKey deactivates an API key.
func (c *Client) RevokeKey(ctx context.Context, tenantID, keyID string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/tenants/"+tenantID+"/keys/"+keyID, nil, nil)
}

// do executes a request, attaching the bearer token and decoding JSON into out.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return ErrUnauthorized
	case resp.StatusCode >= 400:
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("%s %s: %s: %s", method, path, resp.Status, strings.TrimSpace(string(msg)))
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/chaitanyabankanhal/ai-gateway/config"
	"github.com/chaitanyabankanhal/ai-gateway/internal/auth"
	"github.com/chaitanyabankanhal/ai-gateway/internal/db"
	goidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

const stateTTL = 10 * time.Minute

// Service exposes HTTP handlers for the OIDC login flow.
type Service struct {
	cfg      config.OIDCConfig
	admin    config.AdminConfig
	registry *Registry
	state    StateStore
	db       *gorm.DB
}

func NewService(cfg config.OIDCConfig, admin config.AdminConfig, reg *Registry, state StateStore, database *gorm.DB) *Service {
	return &Service{cfg: cfg, admin: admin, registry: reg, state: state, db: database}
}

// Routes mounts /providers, /{provider}/login, /{provider}/callback under the caller's path.
func (s *Service) Routes(r chi.Router) {
	r.Get("/providers", s.handleProviders)
	r.Get("/{provider}/login", s.handleLogin)
	r.Get("/{provider}/callback", s.handleCallback)
}

type providerInfo struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

func (s *Service) handleProviders(w http.ResponseWriter, r *http.Request) {
	names := s.registry.Enabled()
	out := make([]providerInfo, 0, len(names))
	for _, n := range names {
		p, _ := s.registry.Get(n)
		out = append(out, providerInfo{Name: p.Name, Label: p.Label})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (s *Service) handleLogin(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "provider")
	p, ok := s.registry.Get(name)
	if !ok {
		http.Error(w, "unknown provider", http.StatusNotFound)
		return
	}

	state, err := randB64(32)
	if err != nil {
		http.Error(w, "state gen", http.StatusInternalServerError)
		return
	}
	verifier, err := randB64(48)
	if err != nil {
		http.Error(w, "verifier gen", http.StatusInternalServerError)
		return
	}
	nonce, err := randB64(24)
	if err != nil {
		http.Error(w, "nonce gen", http.StatusInternalServerError)
		return
	}

	rec := StateRecord{
		Provider:     name,
		CodeVerifier: verifier,
		Nonce:        nonce,
		CreatedAt:    time.Now().Unix(),
	}
	// Optional CLI loopback redirect. Only loopback URLs are accepted — an
	// open redirect here would let an attacker exfiltrate the minted JWT.
	if cliRedirect := r.URL.Query().Get("cli_redirect"); cliRedirect != "" {
		if !isLoopbackURL(cliRedirect) {
			http.Error(w, "cli_redirect must be a loopback URL", http.StatusBadRequest)
			return
		}
		rec.CLIRedirect = cliRedirect
	}
	if err := s.state.Put(r.Context(), state, rec, stateTTL); err != nil {
		slog.Error("oidc: state put", "err", err)
		http.Error(w, "state store", http.StatusInternalServerError)
		return
	}

	authURL := p.OAuth2.AuthCodeURL(state,
		oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		goidc.Nonce(nonce),
	)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Service) handleCallback(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "provider")

	// cliRedirect is unknown until state is loaded; early errors fall back to
	// the web frontend, which is the only safe default before state is trusted.
	cliRedirect := ""

	p, ok := s.registry.Get(name)
	if !ok {
		s.redirectErr(w, r, cliRedirect, "unknown_provider")
		return
	}

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		slog.Warn("oidc: idp returned error", "provider", name, "error", errParam, "desc", r.URL.Query().Get("error_description"))
		s.redirectErr(w, r, cliRedirect, errParam)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		s.redirectErr(w, r, cliRedirect, "missing_params")
		return
	}

	rec, err := s.state.Take(r.Context(), state)
	if err != nil {
		s.redirectErr(w, r, cliRedirect, "invalid_state")
		return
	}
	// State is now trusted — honor its CLI redirect for all subsequent errors.
	cliRedirect = rec.CLIRedirect
	if rec.Provider != name {
		s.redirectErr(w, r, cliRedirect, "provider_mismatch")
		return
	}

	token, err := p.OAuth2.Exchange(r.Context(), code,
		oauth2.SetAuthURLParam("code_verifier", rec.CodeVerifier),
	)
	if err != nil {
		slog.Warn("oidc: code exchange failed", "provider", name, "err", err)
		s.redirectErr(w, r, cliRedirect, "exchange_failed")
		return
	}

	rawID, ok := token.Extra("id_token").(string)
	if !ok || rawID == "" {
		s.redirectErr(w, r, cliRedirect, "no_id_token")
		return
	}

	idToken, err := p.Verifier.Verify(r.Context(), rawID)
	if err != nil {
		slog.Warn("oidc: id token verify failed", "provider", name, "err", err)
		s.redirectErr(w, r, cliRedirect, "verify_failed")
		return
	}
	if idToken.Nonce != rec.Nonce {
		s.redirectErr(w, r, cliRedirect, "nonce_mismatch")
		return
	}

	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		s.redirectErr(w, r, cliRedirect, "bad_claims")
		return
	}
	email := claims.ResolvedEmail()
	if claims.Subject == "" || email == "" {
		s.redirectErr(w, r, cliRedirect, "missing_identity")
		return
	}

	user, err := s.provisionUser(r.Context(), name, claims.Subject, email, claims.Name)
	if err != nil {
		slog.Error("oidc: provision user", "err", err)
		s.redirectErr(w, r, cliRedirect, "provision_failed")
		return
	}

	ttl := time.Duration(s.admin.JWTTTLHours) * time.Hour
	jwt, err := auth.GenerateToken(user.ID, user.Name, user.Email, user.IsSuperAdmin, s.admin.JWTSecret, ttl)
	if err != nil {
		s.redirectErr(w, r, cliRedirect, "jwt_failed")
		return
	}

	s.redirectSuccess(w, r, cliRedirect, jwt)
}

// provisionUser performs JIT provisioning:
//  1. Existing OIDCIdentity → return its User.
//  2. Existing User by email → link identity (account-merge).
//  3. Else create User w/ empty PasswordHash + identity.
func (s *Service) provisionUser(ctx context.Context, provider, subject, email, name string) (*db.User, error) {
	var ident db.OIDCIdentity
	err := s.db.WithContext(ctx).Where("provider = ? AND subject = ?", provider, subject).First(&ident).Error
	if err == nil {
		var u db.User
		if err := s.db.WithContext(ctx).First(&u, ident.UserID).Error; err != nil {
			return nil, fmt.Errorf("load linked user: %w", err)
		}
		return &u, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("lookup identity: %w", err)
	}

	var user db.User
	err = s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = db.User{Email: email, Name: name}
		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	ident = db.OIDCIdentity{Provider: provider, Subject: subject, UserID: user.ID, Email: email}
	if err := s.db.WithContext(ctx).Create(&ident).Error; err != nil {
		return nil, fmt.Errorf("create identity: %w", err)
	}
	return &user, nil
}

// redirectSuccess sends the browser to the final destination with the JWT.
// For a CLI loopback login the token rides in a query param; for the web
// frontend it rides in the URL fragment (kept out of gateway/proxy logs).
func (s *Service) redirectSuccess(w http.ResponseWriter, r *http.Request, cliRedirect, jwt string) {
	var target string
	if cliRedirect != "" {
		target = appendQuery(cliRedirect, "token", jwt)
	} else {
		target = s.cfg.FrontendURL + "/auth/callback#token=" + url.QueryEscape(jwt)
	}
	http.Redirect(w, r, target, http.StatusFound)
}

// redirectErr sends the browser back with an error reason. When cliRedirect is
// set the reason rides in a query param so the CLI listener can surface it.
func (s *Service) redirectErr(w http.ResponseWriter, r *http.Request, cliRedirect, reason string) {
	var target string
	if cliRedirect != "" {
		target = appendQuery(cliRedirect, "error", reason)
	} else {
		target = s.cfg.FrontendURL + "/auth/callback#error=" + url.QueryEscape(reason)
	}
	http.Redirect(w, r, target, http.StatusFound)
}

// appendQuery adds key=value to a URL's query string, preserving existing params.
func appendQuery(rawURL, key, value string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}

// isLoopbackURL reports whether rawURL is an http(s) URL whose host resolves to
// a loopback address. Used to gate the CLI redirect against open-redirect abuse.
func isLoopbackURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	host := u.Hostname()
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

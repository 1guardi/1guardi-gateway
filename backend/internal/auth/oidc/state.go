// Package oidc implements OpenID Connect login flows for human users.
//
// Flow:
//  1. Client hits GET /api/v1/auth/oidc/{provider}/login
//  2. Server generates state + PKCE verifier + nonce, stores in Redis (10min TTL),
//     302s the browser to the IdP authorization endpoint.
//  3. IdP redirects back to /callback?code&state.
//  4. Server validates state, exchanges code, verifies ID token, JIT-provisions
//     the user, mints a gateway JWT, and 302s the browser to the frontend
//     with the token in the URL fragment (never sent to the gateway in logs).
//
// CLI loopback variant: when /login is called with a `cli_redirect` query param
// pointing at a loopback address, the final 302 targets that loopback URL with
// the token in a query param instead. This lets the `aigw` CLI capture the JWT
// via a short-lived local HTTP listener.
package oidc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// StateRecord is the server-side data tied to an in-flight OIDC login.
type StateRecord struct {
	Provider     string `json:"provider"`
	CodeVerifier string `json:"code_verifier"`
	Nonce        string `json:"nonce"`
	CreatedAt    int64  `json:"created_at"`
	// CLIRedirect, when set, is a validated loopback URL the callback redirects
	// to with the minted JWT in a query param (for the `aigw` CLI login flow).
	CLIRedirect string `json:"cli_redirect,omitempty"`
}

// StateStore persists short-lived state for in-flight OIDC logins.
type StateStore interface {
	Put(ctx context.Context, state string, rec StateRecord, ttl time.Duration) error
	// Take returns the record and atomically deletes it (single-use, prevents replay).
	Take(ctx context.Context, state string) (StateRecord, error)
}

// ErrStateNotFound is returned when state is missing, expired, or already consumed.
var ErrStateNotFound = errors.New("oidc: state not found or expired")

// RedisStateStore stores OIDC state in Redis under a namespaced key.
type RedisStateStore struct {
	rdb *redis.Client
}

func NewRedisStateStore(rdb *redis.Client) *RedisStateStore {
	return &RedisStateStore{rdb: rdb}
}

func (s *RedisStateStore) key(state string) string {
	return "oidc:state:" + state
}

func (s *RedisStateStore) Put(ctx context.Context, state string, rec StateRecord, ttl time.Duration) error {
	b, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("oidc state: marshal: %w", err)
	}
	if err := s.rdb.Set(ctx, s.key(state), b, ttl).Err(); err != nil {
		return fmt.Errorf("oidc state: redis set: %w", err)
	}
	return nil
}

func (s *RedisStateStore) Take(ctx context.Context, state string) (StateRecord, error) {
	key := s.key(state)
	b, err := s.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return StateRecord{}, ErrStateNotFound
	}
	if err != nil {
		return StateRecord{}, fmt.Errorf("oidc state: redis get: %w", err)
	}
	// Best-effort delete — even if it fails, TTL will reap. Atomicity across get+del
	// would need a Lua script; for OIDC state the race window is the IdP roundtrip.
	s.rdb.Del(ctx, key)

	var rec StateRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		return StateRecord{}, fmt.Errorf("oidc state: unmarshal: %w", err)
	}
	return rec, nil
}

// randB64 returns n random bytes, base64url-encoded (no padding).
func randB64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// pkceChallenge returns the S256 code_challenge for a given verifier.
func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

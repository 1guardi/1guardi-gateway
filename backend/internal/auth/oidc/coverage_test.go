package oidc

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/chaitanyabankanhal/ai-gateway/config"
	goidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// --- test doubles ------------------------------------------------------------

// fakeStateStore is an in-memory StateStore for handler tests.
type fakeStateStore struct {
	mu sync.Mutex
	m  map[string]StateRecord
}

func newFakeStore() *fakeStateStore { return &fakeStateStore{m: map[string]StateRecord{}} }

func (f *fakeStateStore) Put(_ context.Context, state string, rec StateRecord, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.m[state] = rec
	return nil
}

func (f *fakeStateStore) Take(_ context.Context, state string) (StateRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec, ok := f.m[state]
	if !ok {
		return StateRecord{}, ErrStateNotFound
	}
	delete(f.m, state)
	return rec, nil
}

// only returns the single record in the store (handler picks a random state key).
func (f *fakeStateStore) only() StateRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, v := range f.m {
		return v
	}
	return StateRecord{}
}

// withProvider attaches a chi route param so unexported handlers can be called directly.
func withProvider(r *http.Request, provider string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider", provider)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

const (
	testIssuer   = "https://idp.test"
	testClientID = "test-client"
)

// signedIDToken mints an RS256 ID token for the verifier built in newTestProvider.
func signedIDToken(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	s, err := tok.SignedString(key)
	require.NoError(t, err)
	return s
}

// newTestProvider builds a Provider whose token endpoint is a local httptest
// server returning idToken, and whose verifier trusts key's public half.
func newTestProvider(t *testing.T, key *rsa.PrivateKey, idToken string, tokenStatus int) *Provider {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if tokenStatus != http.StatusOK {
			w.WriteHeader(tokenStatus)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
			return
		}
		resp := map[string]any{"access_token": "at", "token_type": "Bearer", "expires_in": 3600}
		if idToken != "" {
			resp["id_token"] = idToken
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(ts.Close)

	verifier := goidc.NewVerifier(testIssuer,
		&goidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&key.PublicKey}},
		&goidc.Config{ClientID: testClientID})

	return &Provider{
		Name:  "test",
		Label: "Test",
		OAuth2: &oauth2.Config{
			ClientID:     testClientID,
			ClientSecret: "secret",
			Endpoint:     oauth2.Endpoint{AuthURL: "https://idp.test/authorize", TokenURL: ts.URL},
			RedirectURL:  "http://localhost:8081/api/v1/auth/oidc/test/callback",
			Scopes:       []string{goidc.ScopeOpenID, "email"},
		},
		Verifier: verifier,
	}
}

func newTestService(t *testing.T, p *Provider, store StateStore, database *gorm.DB) *Service {
	t.Helper()
	reg := &Registry{providers: map[string]*Provider{}}
	if p != nil {
		reg.providers[p.Name] = p
	}
	cfg := config.OIDCConfig{FrontendURL: "http://frontend.test"}
	admin := config.AdminConfig{JWTSecret: "unit-test-secret", JWTTTLHours: 1}
	return NewService(cfg, admin, reg, store, database)
}

// --- provider.go -------------------------------------------------------------

func TestRegistry_GetAndEnabled(t *testing.T) {
	reg := &Registry{providers: map[string]*Provider{
		"google":    {Name: "google", Label: "Google"},
		"microsoft": {Name: "microsoft", Label: "Microsoft"},
	}}
	assert.Equal(t, []string{"google", "microsoft"}, reg.Enabled())

	p, ok := reg.Get("google")
	assert.True(t, ok)
	assert.Equal(t, "Google", p.Label)

	_, ok = reg.Get("missing")
	assert.False(t, ok)
}

func TestNewRegistry_NoProvidersEnabled(t *testing.T) {
	reg, err := NewRegistry(context.Background(), config.OIDCConfig{})
	require.NoError(t, err)
	assert.Empty(t, reg.Enabled())
}

func TestRedirectURL(t *testing.T) {
	assert.Equal(t, "http://x/api/v1/auth/oidc/google/callback", redirectURL("http://x", "google"))
}

func TestClaims_ResolvedEmail(t *testing.T) {
	assert.Equal(t, "a@b.com", Claims{Email: "a@b.com"}.ResolvedEmail())
	assert.Equal(t, "u@b.com", Claims{PreferredUsername: "u@b.com"}.ResolvedEmail())
	assert.Equal(t, "", Claims{}.ResolvedEmail())
}

// --- handleProviders / Routes ------------------------------------------------

func TestHandleProviders(t *testing.T) {
	svc := newTestService(t, &Provider{Name: "test", Label: "Test"}, newFakeStore(), nil)

	r := chi.NewRouter()
	r.Route("/auth/oidc", svc.Routes)
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/auth/oidc/providers")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var out []providerInfo
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	require.Len(t, out, 1)
	assert.Equal(t, "test", out[0].Name)
}

// --- handleLogin -------------------------------------------------------------

func TestHandleLogin_Success(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(t, newTestProvider(t, mustKey(t), "", http.StatusOK), store, nil)

	rr := httptest.NewRecorder()
	svc.handleLogin(rr, withProvider(httptest.NewRequest(http.MethodGet, "/test/login", nil), "test"))

	assert.Equal(t, http.StatusFound, rr.Code)
	loc := rr.Header().Get("Location")
	assert.Contains(t, loc, "code_challenge")
	assert.Contains(t, loc, "code_challenge_method=S256")
	assert.NotEmpty(t, store.only().CodeVerifier)
}

func TestHandleLogin_UnknownProvider(t *testing.T) {
	svc := newTestService(t, nil, newFakeStore(), nil)
	rr := httptest.NewRecorder()
	svc.handleLogin(rr, withProvider(httptest.NewRequest(http.MethodGet, "/x/login", nil), "x"))
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleLogin_CLIRedirectStored(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(t, newTestProvider(t, mustKey(t), "", http.StatusOK), store, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test/login?cli_redirect=http://127.0.0.1:5555/callback", nil)
	svc.handleLogin(rr, withProvider(req, "test"))

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "http://127.0.0.1:5555/callback", store.only().CLIRedirect)
}

func TestHandleLogin_CLIRedirectRejectsExternal(t *testing.T) {
	svc := newTestService(t, newTestProvider(t, mustKey(t), "", http.StatusOK), newFakeStore(), nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test/login?cli_redirect=http://evil.com/steal", nil)
	svc.handleLogin(rr, withProvider(req, "test"))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- handleCallback ----------------------------------------------------------

func TestHandleCallback_UnknownProvider(t *testing.T) {
	svc := newTestService(t, nil, newFakeStore(), nil)
	rr := httptest.NewRecorder()
	svc.handleCallback(rr, withProvider(httptest.NewRequest(http.MethodGet, "/x/callback", nil), "x"))
	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "error=unknown_provider")
}

func TestHandleCallback_IdPError(t *testing.T) {
	svc := newTestService(t, newTestProvider(t, mustKey(t), "", http.StatusOK), newFakeStore(), nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test/callback?error=access_denied", nil)
	svc.handleCallback(rr, withProvider(req, "test"))
	assert.Contains(t, rr.Header().Get("Location"), "error=access_denied")
}

func TestHandleCallback_MissingParams(t *testing.T) {
	svc := newTestService(t, newTestProvider(t, mustKey(t), "", http.StatusOK), newFakeStore(), nil)
	rr := httptest.NewRecorder()
	svc.handleCallback(rr, withProvider(httptest.NewRequest(http.MethodGet, "/test/callback", nil), "test"))
	assert.Contains(t, rr.Header().Get("Location"), "error=missing_params")
}

func TestHandleCallback_InvalidState(t *testing.T) {
	svc := newTestService(t, newTestProvider(t, mustKey(t), "", http.StatusOK), newFakeStore(), nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test/callback?state=nope&code=c", nil)
	svc.handleCallback(rr, withProvider(req, "test"))
	assert.Contains(t, rr.Header().Get("Location"), "error=invalid_state")
}

func TestHandleCallback_ProviderMismatch(t *testing.T) {
	store := newFakeStore()
	_ = store.Put(context.Background(), "ST", StateRecord{Provider: "other"}, time.Minute)
	svc := newTestService(t, newTestProvider(t, mustKey(t), "", http.StatusOK), store, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test/callback?state=ST&code=c", nil)
	svc.handleCallback(rr, withProvider(req, "test"))
	assert.Contains(t, rr.Header().Get("Location"), "error=provider_mismatch")
}

func TestHandleCallback_ExchangeFailed(t *testing.T) {
	store := newFakeStore()
	_ = store.Put(context.Background(), "ST", StateRecord{Provider: "test", Nonce: "n"}, time.Minute)
	svc := newTestService(t, newTestProvider(t, mustKey(t), "", http.StatusBadRequest), store, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test/callback?state=ST&code=bad", nil)
	svc.handleCallback(rr, withProvider(req, "test"))
	assert.Contains(t, rr.Header().Get("Location"), "error=exchange_failed")
}

func TestHandleCallback_NoIDToken(t *testing.T) {
	store := newFakeStore()
	_ = store.Put(context.Background(), "ST", StateRecord{Provider: "test", Nonce: "n"}, time.Minute)
	svc := newTestService(t, newTestProvider(t, mustKey(t), "", http.StatusOK), store, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test/callback?state=ST&code=c", nil)
	svc.handleCallback(rr, withProvider(req, "test"))
	assert.Contains(t, rr.Header().Get("Location"), "error=no_id_token")
}

func TestHandleCallback_NonceMismatch(t *testing.T) {
	key := mustKey(t)
	store := newFakeStore()
	_ = store.Put(context.Background(), "ST", StateRecord{Provider: "test", Nonce: "expected"}, time.Minute)
	idTok := signedIDToken(t, key, jwt.MapClaims{
		"iss": testIssuer, "aud": testClientID, "sub": "s1", "email": "a@b.com",
		"nonce": "WRONG", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
	})
	svc := newTestService(t, newTestProvider(t, key, idTok, http.StatusOK), store, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test/callback?state=ST&code=c", nil)
	svc.handleCallback(rr, withProvider(req, "test"))
	assert.Contains(t, rr.Header().Get("Location"), "error=nonce_mismatch")
}

func TestHandleCallback_Success(t *testing.T) {
	key := mustKey(t)
	store := newFakeStore()
	_ = store.Put(context.Background(), "ST", StateRecord{Provider: "test", Nonce: "nonce-1"}, time.Minute)
	idTok := signedIDToken(t, key, jwt.MapClaims{
		"iss": testIssuer, "aud": testClientID, "sub": "sub-xyz", "email": "alice@example.com",
		"name": "Alice", "nonce": "nonce-1", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
	})
	svc := newTestService(t, newTestProvider(t, key, idTok, http.StatusOK), store, newTestDB(t))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test/callback?state=ST&code=c", nil)
	svc.handleCallback(rr, withProvider(req, "test"))

	assert.Equal(t, http.StatusFound, rr.Code)
	loc := rr.Header().Get("Location")
	assert.True(t, strings.HasPrefix(loc, "http://frontend.test/auth/callback#token="), "got %s", loc)
}

func TestHandleCallback_SuccessCLIRedirect(t *testing.T) {
	key := mustKey(t)
	store := newFakeStore()
	_ = store.Put(context.Background(), "ST", StateRecord{
		Provider: "test", Nonce: "nonce-1", CLIRedirect: "http://127.0.0.1:9000/callback",
	}, time.Minute)
	idTok := signedIDToken(t, key, jwt.MapClaims{
		"iss": testIssuer, "aud": testClientID, "sub": "sub-cli", "email": "cli@example.com",
		"name": "CLI User", "nonce": "nonce-1", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
	})
	svc := newTestService(t, newTestProvider(t, key, idTok, http.StatusOK), store, newTestDB(t))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test/callback?state=ST&code=c", nil)
	svc.handleCallback(rr, withProvider(req, "test"))

	loc := rr.Header().Get("Location")
	assert.True(t, strings.HasPrefix(loc, "http://127.0.0.1:9000/callback?token="), "got %s", loc)
}

// --- state.go: RedisStateStore ----------------------------------------------

func TestRedisStateStore_PutTakeSingleUse(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	store := NewRedisStateStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	ctx := context.Background()

	rec := StateRecord{Provider: "google", Nonce: "n", CodeVerifier: "v", CreatedAt: 1}
	require.NoError(t, store.Put(ctx, "abc", rec, time.Minute))

	got, err := store.Take(ctx, "abc")
	require.NoError(t, err)
	assert.Equal(t, rec, got)

	// State is single-use: a second Take must fail.
	_, err = store.Take(ctx, "abc")
	assert.ErrorIs(t, err, ErrStateNotFound)
}

func TestRedisStateStore_TakeMissing(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	store := NewRedisStateStore(redis.NewClient(&redis.Options{Addr: mr.Addr()}))
	_, err = store.Take(context.Background(), "never-stored")
	assert.ErrorIs(t, err, ErrStateNotFound)
}

func mustKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

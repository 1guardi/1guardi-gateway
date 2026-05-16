package oidc

import (
	"context"
	"net/url"
	"testing"

	"github.com/chaitanyabankanhal/ai-gateway/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(database))
	return database
}

func newSvc(database *gorm.DB) *Service {
	// nil registry / state — provisionUser doesn't touch them.
	return &Service{db: database}
}

func TestProvisionUser_NewUserCreated(t *testing.T) {
	database := newTestDB(t)
	svc := newSvc(database)

	user, err := svc.provisionUser(context.Background(), "google", "sub-1", "alice@example.com", "Alice")
	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.Equal(t, "Alice", user.Name)
	assert.Empty(t, user.PasswordHash, "SSO-provisioned user must have no password")

	var ident db.OIDCIdentity
	require.NoError(t, database.Where("provider = ? AND subject = ?", "google", "sub-1").First(&ident).Error)
	assert.Equal(t, user.ID, ident.UserID)
}

func TestProvisionUser_ExistingIdentityReturnsSameUser(t *testing.T) {
	database := newTestDB(t)
	svc := newSvc(database)

	u1, err := svc.provisionUser(context.Background(), "google", "sub-1", "alice@example.com", "Alice")
	require.NoError(t, err)

	u2, err := svc.provisionUser(context.Background(), "google", "sub-1", "alice@example.com", "Alice")
	require.NoError(t, err)
	assert.Equal(t, u1.ID, u2.ID)

	var count int64
	database.Model(&db.OIDCIdentity{}).Count(&count)
	assert.Equal(t, int64(1), count, "no duplicate identity")
}

func TestProvisionUser_LinksToExistingEmail(t *testing.T) {
	// Pre-existing local-auth user; logging in via SSO with matching email should link, not duplicate.
	database := newTestDB(t)
	existing := db.User{Email: "bob@example.com", Name: "Bob", PasswordHash: "legacy-hash"}
	require.NoError(t, database.Create(&existing).Error)

	svc := newSvc(database)
	u, err := svc.provisionUser(context.Background(), "microsoft", "ms-sub-1", "bob@example.com", "Bob")
	require.NoError(t, err)
	assert.Equal(t, existing.ID, u.ID)
	assert.Equal(t, "legacy-hash", u.PasswordHash, "linking must not clobber existing password")

	var count int64
	database.Model(&db.User{}).Where("email = ?", "bob@example.com").Count(&count)
	assert.Equal(t, int64(1), count, "no duplicate user")
}

func TestIsLoopbackURL(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"http://127.0.0.1:54321/callback", true},
		{"http://localhost:8080/callback?cli_state=abc", true},
		{"http://[::1]:9000/callback", true},
		{"https://127.0.0.1/callback", true},
		{"http://evil.com/callback", false},
		{"http://169.254.169.254/callback", false},
		{"http://10.0.0.5/callback", false},
		{"ftp://127.0.0.1/callback", false},
		{"javascript:alert(1)", false},
		{"not a url at all %%", false},
		{"", false},
	}
	for _, c := range cases {
		assert.Equalf(t, c.want, isLoopbackURL(c.url), "isLoopbackURL(%q)", c.url)
	}
}

func TestAppendQuery(t *testing.T) {
	got := appendQuery("http://127.0.0.1:5000/callback?cli_state=xyz", "token", "jwt.value.here")
	u, err := url.Parse(got)
	require.NoError(t, err)
	assert.Equal(t, "xyz", u.Query().Get("cli_state"), "existing param preserved")
	assert.Equal(t, "jwt.value.here", u.Query().Get("token"), "new param added")
}

func TestProvisionUser_DifferentProvidersDifferentIdentities(t *testing.T) {
	database := newTestDB(t)
	svc := newSvc(database)

	u1, err := svc.provisionUser(context.Background(), "google", "sub-1", "alice@example.com", "Alice")
	require.NoError(t, err)
	u2, err := svc.provisionUser(context.Background(), "microsoft", "sub-1", "alice@example.com", "Alice")
	require.NoError(t, err)

	// Same email → same user, but two separate identities.
	assert.Equal(t, u1.ID, u2.ID)

	var count int64
	database.Model(&db.OIDCIdentity{}).Count(&count)
	assert.Equal(t, int64(2), count)
}

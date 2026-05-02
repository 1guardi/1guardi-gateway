package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	secret := "test-secret"
	ttl := time.Hour

	t.Run("GenerateAndValidate", func(t *testing.T) {
		token, err := GenerateToken(1, "Test User", "test@example.com", true, secret, ttl)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := ValidateToken(token, secret)
		require.NoError(t, err)
		assert.Equal(t, uint(1), claims.UserID)
		assert.Equal(t, "Test User", claims.Name)
		assert.Equal(t, "test@example.com", claims.Email)
		assert.True(t, claims.IsSuperAdmin)
	})

	t.Run("InvalidSecret", func(t *testing.T) {
		token, _ := GenerateToken(1, "User", "u@e.com", false, secret, ttl)
		_, err := ValidateToken(token, "wrong-secret")
		assert.Error(t, err)
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		token, _ := GenerateToken(1, "User", "u@e.com", false, secret, -time.Hour)
		_, err := ValidateToken(token, secret)
		assert.Error(t, err)
	})
}

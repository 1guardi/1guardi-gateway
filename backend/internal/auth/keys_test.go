package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAPIKey(t *testing.T) {
	key, hash, err := GenerateAPIKey()
	assert.NoError(t, err)

	// Verify key format
	assert.True(t, strings.HasPrefix(key, KeyPrefix+"_"), "key should have correct prefix")

	// Prefix 'sk_' (3 chars) + 32 bytes hex encoded (64 chars) = 67 chars total
	assert.Equal(t, 3+KeyLength*2, len(key), "key should have correct length")

	// Verify hash format
	assert.Equal(t, 64, len(hash), "hash should be a 64-character hex string")

	// Verify that the hash matches the actual SHA-256 of the key
	expectedHash := HashKey(key)
	assert.Equal(t, expectedHash, hash, "returned hash should match HashKey output")
}

func TestGenerateAPIKey_Randomness(t *testing.T) {
	key1, _, err := GenerateAPIKey()
	assert.NoError(t, err)

	key2, _, err := GenerateAPIKey()
	assert.NoError(t, err)

	assert.NotEqual(t, key1, key2, "generated keys should be unique")
}

func TestHashKey(t *testing.T) {
	key := "sk_test_1234567890"

	// Calculate expected hash manually
	h := sha256.Sum256([]byte(key))
	expected := hex.EncodeToString(h[:])

	hash := HashKey(key)
	assert.Equal(t, expected, hash)
}

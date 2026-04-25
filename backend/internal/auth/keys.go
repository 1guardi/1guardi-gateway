package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	KeyPrefix = "sk"
	KeyLength = 32
)

// GenerateAPIKey generates a new API key and its hash.
// Format: sk_[random_hex]
func GenerateAPIKey() (string, string, error) {
	bytes := make([]byte, KeyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	key := fmt.Sprintf("%s_%s", KeyPrefix, hex.EncodeToString(bytes))
	hash := HashKey(key)

	return key, hash, nil
}

// HashKey returns the SHA-256 hash of the API key.
func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

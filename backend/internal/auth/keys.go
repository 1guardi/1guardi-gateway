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

// GenerateAPIKey generates a new API key, its hash, and its suffix.
// Format: sk_[random_hex]
func GenerateAPIKey() (key string, hash string, suffix string, err error) {
	bytes := make([]byte, KeyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", "", err
	}

	key = fmt.Sprintf("%s_%s", KeyPrefix, hex.EncodeToString(bytes))
	hash = HashKey(key)
	suffix = key[len(key)-4:]

	return key, hash, suffix, nil
}

// HashKey returns the SHA-256 hash of the API key.
func HashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

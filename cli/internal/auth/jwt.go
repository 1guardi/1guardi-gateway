package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Identity is the subset of gateway JWT claims the CLI surfaces to the user.
type Identity struct {
	UserID       uint
	Name         string
	Email        string
	IsSuperAdmin bool
	Expiry       time.Time
}

// DecodeIdentity parses a gateway JWT WITHOUT verifying its signature — the
// CLI has no signing key and only uses the claims for display. The gateway
// re-verifies every token, so unverified local decode is safe for UX only.
func DecodeIdentity(token string) (*Identity, error) {
	var claims jwt.MapClaims
	parser := jwt.NewParser()
	if _, _, err := parser.ParseUnverified(token, &claims); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}

	id := &Identity{}
	if v, ok := claims["uid"].(float64); ok {
		id.UserID = uint(v)
	}
	id.Name, _ = claims["name"].(string)
	id.Email, _ = claims["email"].(string)
	id.IsSuperAdmin, _ = claims["is_super_admin"].(bool)
	if exp, err := claims.GetExpirationTime(); err == nil && exp != nil {
		id.Expiry = exp.Time
	}
	return id, nil
}

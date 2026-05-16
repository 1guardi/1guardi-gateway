package oidc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPKCEChallenge_S256_KnownVector(t *testing.T) {
	// From RFC 7636 Appendix B.
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	assert.Equal(t, want, pkceChallenge(verifier))
}

func TestRandB64_LengthAndUniqueness(t *testing.T) {
	a, err := randB64(32)
	assert.NoError(t, err)
	b, err := randB64(32)
	assert.NoError(t, err)
	assert.NotEqual(t, a, b)
	// 32 bytes → 43 chars base64url (no padding).
	assert.Len(t, a, 43)
}

func TestErrStateNotFound_Identity(t *testing.T) {
	assert.True(t, errors.Is(ErrStateNotFound, ErrStateNotFound))
}

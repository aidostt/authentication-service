package authManager

import (
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	m, err := NewManager("test-signing-key")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return m
}

func TestNewManager_EmptyKeyRejected(t *testing.T) {
	if _, err := NewManager(""); err == nil {
		t.Fatal("expected an error for an empty signing key, got nil")
	}
}

func TestManager_AccessTokenRoundTrip(t *testing.T) {
	m := newTestManager(t)
	wantRoles := []string{"user", "admin"}

	token, err := m.NewAccessToken("user-1", time.Hour, wantRoles, "reservista", true)
	if err != nil {
		t.Fatalf("NewAccessToken: %v", err)
	}

	id, roles, activated, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if id != "user-1" {
		t.Errorf("id = %q, want %q", id, "user-1")
	}
	if !activated {
		t.Error("activated = false, want true")
	}
	if len(roles) != len(wantRoles) || roles[0] != "user" || roles[1] != "admin" {
		t.Errorf("roles = %v, want %v", roles, wantRoles)
	}
}

func TestManager_ParseExpiredReturnsSentinelAndClaims(t *testing.T) {
	m := newTestManager(t)

	// A negative TTL yields an already-expired token.
	token, err := m.NewAccessToken("user-1", -time.Minute, []string{"user"}, "reservista", false)
	if err != nil {
		t.Fatalf("NewAccessToken: %v", err)
	}

	id, _, _, err := m.Parse(token)
	if !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("Parse error = %v, want ErrTokenExpired", err)
	}
	if id != "user-1" {
		t.Errorf("expired token should still expose its claims: id = %q, want %q", id, "user-1")
	}
}

func TestManager_ParseRejectsAlgNone(t *testing.T) {
	m := newTestManager(t)

	// Forge a token with alg "none"; algorithm pinning must reject it even
	// though it carries otherwise valid claims.
	claims := CustomClaims{
		UserID:           "attacker",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none token: %v", err)
	}

	if _, _, _, err := m.Parse(token); err == nil {
		t.Fatal("expected Parse to reject an alg=none token, got nil error")
	}
}

func TestManager_ParseRejectsWrongKey(t *testing.T) {
	signer := newTestManager(t)
	verifier, err := NewManager("a-different-key")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	token, err := signer.NewAccessToken("user-1", time.Hour, []string{"user"}, "reservista", false)
	if err != nil {
		t.Fatalf("NewAccessToken: %v", err)
	}

	if _, _, _, err := verifier.Parse(token); err == nil {
		t.Fatal("expected Parse to reject a token signed with a different key")
	}
}

func TestManager_NewRefreshTokenIsRandomHex(t *testing.T) {
	m := newTestManager(t)

	first, err := m.NewRefreshToken()
	if err != nil {
		t.Fatalf("NewRefreshToken: %v", err)
	}
	second, err := m.NewRefreshToken()
	if err != nil {
		t.Fatalf("NewRefreshToken: %v", err)
	}

	if first == second {
		t.Fatal("consecutive refresh tokens must differ")
	}
	if len(first) != 64 { // 32 random bytes hex-encoded
		t.Errorf("refresh token length = %d, want 64", len(first))
	}
	if _, err := hex.DecodeString(first); err != nil {
		t.Errorf("refresh token is not valid hex: %v", err)
	}
}

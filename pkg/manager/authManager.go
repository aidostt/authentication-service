package authManager

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// ErrTokenExpired is returned by Parse when the access token is well-formed and
// correctly signed but past its expiration. Callers match it with errors.Is to
// decide whether to refresh, rather than comparing error strings.
var ErrTokenExpired = errors.New("token is expired")

// TokenManager provides logic for JWT & Refresh tokens generation and parsing.
type TokenManager interface {
	NewAccessToken(string, time.Duration, []string, string, bool) (string, error)
	Parse(accessToken string) (string, []string, bool, error)
	NewRefreshToken() (string, error)
	NewActivationToken(string) string
}

type Manager struct {
	signingKey string
}

type CustomClaims struct {
	UserID    string   `json:"user_id"`
	Roles     []string `json:"roles"`
	Activated bool     `json:"activated"`
	jwt.RegisteredClaims
}

func NewManager(signingKey string) (*Manager, error) {
	if signingKey == "" {
		return nil, errors.New("empty signing key")
	}

	return &Manager{signingKey: signingKey}, nil
}

func (m *Manager) NewAccessToken(userID string, ttl time.Duration, roles []string, issuer string, activated bool) (string, error) {
	claims := CustomClaims{
		UserID:    userID,
		Roles:     roles,
		Activated: activated,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(m.signingKey))
}

// Parse verifies the access token's signature and claims and returns the
// identity it carries. The accepted signing algorithm is pinned server-side, so
// a token advertising a different algorithm is rejected. An expired but
// otherwise valid token still yields its claims alongside ErrTokenExpired, so
// the caller can refresh using the (now stale) identity.
func (m *Manager) Parse(accessToken string) (string, []string, bool, error) {
	claims := new(CustomClaims)
	_, err := jwt.ParseWithClaims(accessToken, claims, func(*jwt.Token) (interface{}, error) {
		return []byte(m.signingKey), nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return claims.UserID, claims.Roles, claims.Activated, ErrTokenExpired
		}
		return "", nil, false, fmt.Errorf("parse access token: %w", err)
	}
	return claims.UserID, claims.Roles, claims.Activated, nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (m *Manager) NewActivationToken(data string) string {
	mac := hmac.New(sha256.New, []byte(m.signingKey))
	mac.Write([]byte(data))
	signature := mac.Sum(nil)
	token := fmt.Sprintf("%s.%s", data, base64.URLEncoding.EncodeToString(signature))
	return base64.URLEncoding.EncodeToString([]byte(token))
}

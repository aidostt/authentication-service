package authManager

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"time"
)

// TokenManager provides logic for JWT & Refresh tokens generation and parsing.
type TokenManager interface {
	NewAccessToken(string, time.Duration, []string, string, bool) (string, error)
	Parse(accessToken string) (string, []string, bool, error)
	NewRefreshToken() (string, error)
	HexToObjectID(string) (primitive.ObjectID, error)
	NewActivationToken(string) string
}

type Manager struct {
	signingKey string
}

type CustomClaims struct {
	UserID    string   `json:"user_id"`
	Roles     []string `json:"roles"`
	Activated bool     `json:"activated"`
	jwt.StandardClaims
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
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(), // Token expires in 24 hours
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(m.signingKey))
}

// Parse taking from the payload of JWT user id and returns it in string format. Token is still returned
// in both cases, if it is expired or not.
func (m *Manager) Parse(accessToken string) (string, []string, bool, error) {
	token, err := jwt.ParseWithClaims(accessToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.signingKey), nil
	})

	if err != nil {
		var validationError *jwt.ValidationError
		if errors.As(err, &validationError) && validationError.Errors&jwt.ValidationErrorExpired != 0 {
			err = errors.New("token is expired")
		} else {
			return "", nil, false, err
		}
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", nil, false, fmt.Errorf("error getting user claims from token")
	}

	return claims.UserID, claims.Roles, claims.Activated, err
}

func (m *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}

func (m *Manager) HexToObjectID(hex string) (primitive.ObjectID, error) {
	objectId, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return objectId, nil
}

func (m *Manager) NewActivationToken(data string) string {
	mac := hmac.New(sha256.New, []byte(m.signingKey))
	mac.Write([]byte(data))
	signature := mac.Sum(nil)
	token := fmt.Sprintf("%s.%s", data, base64.URLEncoding.EncodeToString(signature))
	return base64.URLEncoding.EncodeToString([]byte(token))
}

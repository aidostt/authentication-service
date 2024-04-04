package service

import (
	"authentication-service/internal/domain"
	"authentication-service/internal/repository"
	"authentication-service/pkg/hash"
	auth "authentication-service/pkg/manager"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type SessionService struct {
	repo            repository.Sessions
	hasher          hash.PasswordHasher
	tokenManager    auth.TokenManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	domain string
}

func NewSessionsService(repo repository.Sessions, hasher hash.PasswordHasher, tokenManager auth.TokenManager, accessTTL, refreshTTL time.Duration, domain string) *SessionService {
	return &SessionService{
		repo:            repo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		domain:          domain,
	}
}

func (s *SessionService) RefreshTokens(ctx context.Context, accessToken string) (string, error) {
	hex, err := s.tokenManager.Parse(accessToken)
	if err != nil {
		return "", err
	}
	id, err := s.tokenManager.HexToObjectID(hex)
	if err != nil {
		return "", err
	}

	user, err := s.repo.GetByUserID(ctx, id)
	if err != nil {
		return "", err
	}

	return s.CreateSession(ctx, user.ID)
}

func (s *SessionService) CreateSession(ctx context.Context, userId primitive.ObjectID) (string, error) {

	AT, err := s.tokenManager.NewAccessToken(userId.Hex(), s.accessTokenTTL)
	if err != nil {
		return "", err
	}

	RT, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		return "", err
	}

	session := domain.Session{
		UserID:       userId,
		RefreshToken: RT,
		ExpiredAt:    time.Now().Add(s.refreshTokenTTL),
	}

	err = s.repo.SetSession(ctx, session)

	return AT, err
}

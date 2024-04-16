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

func NewSessionService(repo repository.Sessions, hasher hash.PasswordHasher, tokenManager auth.TokenManager, accessTTL, refreshTTL time.Duration, domain string) *SessionService {
	return &SessionService{
		repo:            repo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		domain:          domain,
	}
}

func (s *SessionService) Refresh(ctx context.Context, id primitive.ObjectID) (TokenPair, error) {
	return s.CreateSession(ctx, id)
}

func (s *SessionService) CreateSession(ctx context.Context, userId primitive.ObjectID) (res TokenPair, err error) {
	res.AccessToken, err = s.tokenManager.NewAccessToken(userId.Hex(), s.accessTokenTTL)
	if err != nil {
		return TokenPair{}, err
	}

	res.RefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	session := domain.Session{
		UserID:       userId,
		RefreshToken: res.RefreshToken,
		ExpiredAt:    time.Now().Add(s.refreshTokenTTL),
	}

	err = s.repo.SetSession(ctx, session)

	return
}

func (s *SessionService) GetSession(ctx context.Context, RT string) (*domain.Session, error) {
	session, err := s.repo.GetByRefreshToken(ctx, RT)

	if err != nil {
		return nil, err

	}
	// Check if the session has been retrieved and if it is expired
	if session != nil && session.ExpiredAt.After(time.Now()) {
		// Session is expired, handle accordingly
		return nil, domain.ErrSessionExpired
	}
	return session, nil
}

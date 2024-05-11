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

	domain      string
	application string
}

func NewSessionService(repo repository.Sessions, hasher hash.PasswordHasher, tokenManager auth.TokenManager, accessTTL, refreshTTL time.Duration, domain string, application string) *SessionService {
	return &SessionService{
		repo:            repo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		domain:          domain,
		application:     application,
	}
}

func (s *SessionService) Refresh(ctx context.Context, userID primitive.ObjectID, jwt string) (TokenPair, error) {
	useridJwt, roles, err := s.tokenManager.Parse(jwt)
	if err != nil {
		return TokenPair{}, err
	}
	if useridJwt != userID.Hex() {
		return TokenPair{}, domain.ErrUnathorized
	}

	return s.CreateSession(ctx, userID, roles)
}

func (s *SessionService) CreateSession(ctx context.Context, userID primitive.ObjectID, roles []string) (res TokenPair, err error) {
	res.AccessToken, err = s.tokenManager.NewAccessToken(userID.Hex(), s.accessTokenTTL, roles, s.application)
	if err != nil {
		return TokenPair{}, err
	}

	res.RefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	session := domain.Session{
		UserID:       userID,
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
	if session != nil && session.ExpiredAt.Before(time.Now()) {
		// Session is expired, handle accordingly
		return nil, domain.ErrSessionExpired
	}
	return session, nil
}

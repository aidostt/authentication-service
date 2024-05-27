package service

import (
	"authentication-service/internal/domain"
	"authentication-service/internal/repository"
	"authentication-service/pkg/hash"
	auth "authentication-service/pkg/manager"
	"context"
	"fmt"
	"time"
)

type SessionService struct {
	repo               repository.Sessions
	hasher             hash.PasswordHasher
	tokenManager       auth.TokenManager
	activationTokenTTL time.Duration
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration

	domain      string
	application string
}

func NewSessionService(repo repository.Sessions, hasher hash.PasswordHasher, tokenManager auth.TokenManager, accessTTL, refreshTTL, activationTTL time.Duration, domain string, application string) *SessionService {
	return &SessionService{
		repo:               repo,
		hasher:             hasher,
		tokenManager:       tokenManager,
		accessTokenTTL:     accessTTL,
		refreshTokenTTL:    refreshTTL,
		activationTokenTTL: activationTTL,
		domain:             domain,
		application:        application,
	}
}

func (s *SessionService) Refresh(ctx context.Context, user *domain.User, jwt string) (TokenPair, error) {
	useridJwt, _, _, err := s.tokenManager.Parse(jwt)
	if err != nil {
		if err.Error() == "token is expired" {
		} else {
			return TokenPair{}, err
		}
	}
	if useridJwt != user.ID.Hex() {
		return TokenPair{}, domain.ErrUnauthorized
	}
	return s.CreateSession(ctx, user.ID.Hex(), user.Roles, user.Activated)
}

func (s *SessionService) CreateSession(ctx context.Context, userID string, roles []string, activated bool) (res TokenPair, err error) {
	res.AccessToken, err = s.tokenManager.NewAccessToken(userID, s.accessTokenTTL, roles, s.application, activated)
	if err != nil {
		return TokenPair{}, err
	}

	res.RefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}
	id, err := s.tokenManager.HexToObjectID(userID)
	if err != nil {
		return TokenPair{}, err
	}

	session := domain.Session{
		UserID:       id,
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

func (s *SessionService) CreateActivationToken(ctx context.Context, id string) string {
	return s.tokenManager.NewActivationToken(fmt.Sprintf("%v:%v", id, time.Now().Add(s.activationTokenTTL).Unix()))
}

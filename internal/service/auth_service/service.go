package auth_service

import (
	"authentication-service/internal/repository"
	"authentication-service/pkg/hash"
	authManager "authentication-service/pkg/manager"
	"context"
	"github.com/aidostt/protos/gen/go/reservista"
	"time"
)

type AuthService struct {
	repo            repository.Users
	hasher          hash.PasswordHasher
	tokenManager    authManager.TokenManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	domain          string
}

func NewAuthService(repo repository.Users, hasher hash.PasswordHasher, tokenManager authManager.TokenManager, accessTTL, refreshTTL time.Duration, domain string) *AuthService {
	return &AuthService{
		repo:            repo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		domain:          domain,
	}
}

func (s *AuthService) SignUp(context.Context, *reservista.RegisterRequest) (*reservista.TokenResponse, error) {
	return nil, nil
}
func (s *AuthService) SignIn(context.Context, *reservista.SignInRequest) (*reservista.TokenResponse, error) {
	return nil, nil
}
func (s *AuthService) Refresh(context.Context, *reservista.TokenRequest) (*reservista.TokenResponse, error) {
	session, err := s.repo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return TokenPair{}, err
	}

	return s.CreateSession(ctx, session.UserID)
}
func (s *AuthService) CreateSession(context.Context, *reservista.IsAdminRequest) (*reservista.TokenResponse, error) {
	return nil, nil
}
func (s *AuthService) GetToken(context.Context, string) (string, error) { return "", nil }
func (s *AuthService) IsAdmin(context.Context, *reservista.IsAdminRequest) (*reservista.IsAdminResponse, error) {
	return nil, nil
}
func (s *AuthService) SignOut(ctx context.Context, request *reservista.TokenRequest) (*reservista.SignOutResponse, error) {
	return nil, nil
}

func (s *AuthService) mustEmbedUnimplementedAuthServer() {
	//TODO implement me
	panic("implement me")
}

package service

import (
	"authentication-service/internal/repository"
	"authentication-service/internal/service/auth_service"
	"authentication-service/pkg/hash"
	auth "authentication-service/pkg/manager"
	"context"
	"github.com/aidostt/protos/gen/go/reservista"
	"time"
)

type Authentication interface {
	SignUp(context.Context, *reservista.RegisterRequest) (*reservista.TokenResponse, error)
	SignIn(context.Context, *reservista.SignInRequest) (*reservista.TokenResponse, error)
	Refresh(context.Context, *reservista.TokenRequest) (*reservista.TokenResponse, error)
	CreateSession(context.Context, *reservista.IsAdminRequest) (*reservista.TokenResponse, error)
	GetToken(context.Context, string) (string, error)
	IsAdmin(context.Context, *reservista.IsAdminRequest) (*reservista.IsAdminResponse, error)
	SignOut(ctx context.Context, request *reservista.TokenRequest) (*reservista.SignOutResponse, error)
}

type Services struct {
	Auth Authentication
}

type Dependencies struct {
	Repos           *repository.Models
	Hasher          hash.PasswordHasher
	TokenManager    auth.TokenManager
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Environment     string
	Domain          string
}

func NewServices(deps Dependencies) *Services {
	authService := auth_service.NewAuthService(deps.Repos.Users, deps.Hasher, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.Domain)
	return &Services{
		Auth: authService,
	}
}

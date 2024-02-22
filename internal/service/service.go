package service

import (
	"authentication-service/internal/repository"
	"authentication-service/pkg/hash"
	auth "authentication-service/pkg/manager"
	"context"
	"time"
)

type UserSignUpInput struct {
	Name     string
	Email    string
	Phone    string
	Password string
}

type UserSignInInput struct {
	Email    string
	Password string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type Users interface {
	SignUp(ctx context.Context, input UserSignUpInput) error
	SignIn(ctx context.Context, input UserSignInInput) (Tokens, error)
	RefreshTokens(ctx context.Context, refreshToken string) (Tokens, error)
}

type Services struct {
	Users Users
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
	usersService := NewUsersService(deps.Repos.Users, deps.Hasher, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.Domain)

	return &Services{
		Users: usersService,
	}
}

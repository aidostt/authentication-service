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

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type UserSignUpInput struct {
	Name     string
	Surname  string
	Phone    string
	Email    string
	Password string
}

type UserSignInInput struct {
	Email    string
	Password string
}

type Users interface {
	GetByID(context.Context, string) (*domain.User, error)
	GetByEmail(context.Context, string) (*domain.User, error)
	Update(context.Context, string, string, string, string, string, string, []string, bool) (string, error)
	Delete(context.Context, string, string) error
	SignUp(context.Context, string, string, string, string, string, []string) (primitive.ObjectID, string, error)
	SignIn(context.Context, string, string) (primitive.ObjectID, []string, bool, error)
	IsAdmin(context.Context, string) (bool, error)
	Activate(context.Context, string, bool) error
	GenerateVerificationCode() string
}

type Sessions interface {
	CreateActivationToken(context.Context, string) string
	Refresh(context.Context, *domain.User, string) (TokenPair, error)
	CreateSession(context.Context, string, []string, bool) (TokenPair, error)
	GetSession(context.Context, string) (*domain.Session, error)
}

type Services struct {
	Users    Users
	Sessions Sessions
}

type Dependencies struct {
	Repos              *repository.Models
	Hasher             hash.PasswordHasher
	TokenManager       auth.TokenManager
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	ActivationTokenTTL time.Duration
	ActivationCodeTTL  time.Duration
	Environment        string
	Domain             string
	Application        string
}

func NewServices(deps Dependencies) *Services {
	userService := NewUserService(deps.Repos.Users, deps.Hasher, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.ActivationCodeTTL, deps.Domain, deps.Application)
	sessionService := NewSessionService(deps.Repos.Sessions, deps.Hasher, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.ActivationTokenTTL, deps.Domain, deps.Application)
	return &Services{
		Users:    userService,
		Sessions: sessionService,
	}
}

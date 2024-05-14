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
	Update(context.Context, string, string, string, string, string, string, []string) error
	Delete(context.Context, string, string) error
	SignUp(context.Context, string, string, string, string, string, []string) (primitive.ObjectID, error)
	SignIn(context.Context, string, string) (primitive.ObjectID, []string, error)
	IsAdmin(context.Context, string) (bool, error)
}

type Sessions interface {
	Refresh(context.Context, *domain.User, string) (TokenPair, error)
	CreateSession(context.Context, primitive.ObjectID, []string) (TokenPair, error)
	GetSession(context.Context, string) (*domain.Session, error)
}

type Services struct {
	Users    Users
	Sessions Sessions
}

type Dependencies struct {
	Repos           *repository.Models
	Hasher          hash.PasswordHasher
	TokenManager    auth.TokenManager
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Environment     string
	Domain          string
	Application     string
}

func NewServices(deps Dependencies) *Services {
	userService := NewUserService(deps.Repos.Users, deps.Hasher, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.Domain, deps.Application)
	sessionService := NewSessionService(deps.Repos.Sessions, deps.Hasher, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.Domain, deps.Application)
	return &Services{
		Users:    userService,
		Sessions: sessionService,
	}
}

package service

import (
	"authentication-service/internal/repository"
	"authentication-service/pkg/hash"
	auth "authentication-service/pkg/manager"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

//TODO: refactor tokenPair struct

type Session interface {
	RefreshTokens(context.Context, string) (string, error)
	CreateSession(context.Context, primitive.ObjectID) (string, error)
}

type Users interface {
	SignUp(context.Context, UserSignUpInput) (string, error)
	SignIn(context.Context, UserSignInInput) (string, error)
	CreateSession(context.Context, primitive.ObjectID) (string, error)
}

type Services struct {
	Users   Users
	Session Session
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
	sessionService := NewSessionsService(deps.Repos.Sessions, deps.Hasher, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.Domain)
	usersService := NewUsersService(deps.Repos.Users, deps.Hasher, deps.TokenManager, deps.AccessTokenTTL, deps.RefreshTokenTTL, deps.Domain, sessionService)
	return &Services{
		Users:   usersService,
		Session: sessionService,
	}
}

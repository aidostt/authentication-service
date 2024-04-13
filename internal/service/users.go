package service

import (
	"authentication-service/internal/domain"
	"authentication-service/internal/repository"
	"authentication-service/pkg/hash"
	authManager "authentication-service/pkg/manager"
	"context"
	"errors"
	"github.com/aidostt/protos/gen/go/reservista"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserService struct {
	repo            repository.Users
	hasher          hash.PasswordHasher
	tokenManager    authManager.TokenManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	domain          string
}

func NewUserService(repo repository.Users, hasher hash.PasswordHasher, tokenManager authManager.TokenManager, accessTTL, refreshTTL time.Duration, domain string) *UserService {
	return &UserService{
		repo:            repo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		domain:          domain,
	}
}

func (s *UserService) SignUp(ctx context.Context, name string, surname string, phone string, email string, password string) (primitive.ObjectID, error) {

	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	user := &domain.User{
		Name:     name,
		Surname:  surname,
		Phone:    phone,
		Email:    email,
		Password: passwordHash,
	}
	if err = s.repo.Create(ctx, user); err != nil {
		return primitive.ObjectID{}, err
	}
	return user.ID, nil
}
func (s *UserService) SignIn(ctx context.Context, email string, password string) (primitive.ObjectID, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	//TODO: compare passwords
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return primitive.ObjectID{}, err
		}
		return primitive.ObjectID{}, err
	}

	return user.ID, err
}

func (s *UserService) IsAdmin(ctx context.Context, input *reservista.IsAdminRequest) (bool, error) {
	return false, nil
}

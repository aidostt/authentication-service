package service

import (
	"authentication-service/internal/domain"
	"authentication-service/internal/repository"
	"authentication-service/pkg/hash"
	authManager "authentication-service/pkg/manager"
	"context"
	"errors"
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
		Password: string(passwordHash),
	}
	if err = s.repo.Create(ctx, user); err != nil {
		return primitive.ObjectID{}, err
	}
	return user.ID, nil
}
func (s *UserService) SignIn(ctx context.Context, email string, password string) (primitive.ObjectID, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return primitive.ObjectID{}, err
		}
		return primitive.ObjectID{}, err
	}
	ok, err := s.hasher.Matches(password, []byte(user.Password))
	if err != nil {
		return primitive.ObjectID{}, err
	}
	if !ok {
		return primitive.ObjectID{}, domain.ErrWrongPassword
	}
	return user.ID, err
}

func (s *UserService) IsAdmin(ctx context.Context, userID string) (bool, error) {
	return false, nil
}

func (s *UserService) GetByID(ctx context.Context, userID string) (domain.User, error) {
	id, err := s.tokenManager.HexToObjectID(userID)
	if err != nil {
		return domain.User{}, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *UserService) Update(ctx context.Context, userID, name, surname, phone, email, password string) error {
	id, err := s.tokenManager.HexToObjectID(userID)
	if err != nil {
		return err
	}
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return err
	}
	usr := domain.User{
		ID:       id,
		Name:     name,
		Surname:  surname,
		Phone:    phone,
		Email:    email,
		Password: string(passwordHash)}

	return s.repo.Update(ctx, usr)
}
func (s *UserService) Delete(ctx context.Context, userID, email string) error {
	id, err := s.tokenManager.HexToObjectID(userID)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id, email)
}

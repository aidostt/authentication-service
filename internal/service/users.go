package service

import (
	"authentication-service/internal/domain"
	"authentication-service/internal/repository"
	"authentication-service/pkg/hash"
	authManager "authentication-service/pkg/manager"
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"time"
)

type UserService struct {
	repo              repository.Users
	hasher            hash.PasswordHasher
	tokenManager      authManager.TokenManager
	accessTokenTTL    time.Duration
	refreshTokenTTL   time.Duration
	activationCodeTTL time.Duration
	domain            string
	application       string
}

func NewUserService(repo repository.Users, hasher hash.PasswordHasher, tokenManager authManager.TokenManager, accessTTL, refreshTTL, codeTTL time.Duration, domain string, application string) *UserService {
	return &UserService{
		repo:              repo,
		hasher:            hasher,
		tokenManager:      tokenManager,
		accessTokenTTL:    accessTTL,
		refreshTokenTTL:   refreshTTL,
		activationCodeTTL: codeTTL,
		domain:            domain,
		application:       application,
	}
}

func (s *UserService) SignUp(ctx context.Context, name, surname, phone, email, password string, roles []string) (primitive.ObjectID, string, error) {
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return primitive.ObjectID{}, "", err
	}
	newVerificationCode := domain.VerificationCode{
		Code:      s.GenerateVerificationCode(),
		ExpiredAt: time.Now().Add(s.activationCodeTTL),
	}
	user := &domain.User{
		Name:             name,
		Surname:          surname,
		Phone:            phone,
		Email:            email,
		Roles:            roles,
		Password:         string(passwordHash),
		Activated:        false,
		VerificationCode: newVerificationCode,
	}
	if err = s.repo.Create(ctx, user); err != nil {
		return primitive.ObjectID{}, "", err
	}
	return user.ID, user.VerificationCode.Code, nil
}
func (s *UserService) SignIn(ctx context.Context, email string, password string) (primitive.ObjectID, []string, bool, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return primitive.ObjectID{}, nil, false, err
		}
		return primitive.ObjectID{}, nil, false, err
	}
	ok, err := s.hasher.Matches(password, []byte(user.Password))
	if err != nil {
		return primitive.ObjectID{}, nil, false, err
	}
	if !ok {
		return primitive.ObjectID{}, nil, false, domain.ErrWrongPassword
	}
	return user.ID, user.Roles, user.Activated, err
}

func (s *UserService) IsAdmin(ctx context.Context, userID string) (bool, error) {
	id, err := s.tokenManager.HexToObjectID(userID)
	if err != nil {
		return false, err
	}
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	for _, role := range user.Roles {
		if role == domain.AdminRole {
			return true, nil
		}
	}
	return false, nil
}

func (s *UserService) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	id, err := s.tokenManager.HexToObjectID(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *UserService) Update(ctx context.Context, userID, name, surname, phone, email, password string, roles []string, activated bool) (string, error) {
	id, err := s.tokenManager.HexToObjectID(userID)
	if err != nil {
		return "", err
	}
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return "", err
	}
	newVerificationCode := domain.VerificationCode{
		Code:      s.GenerateVerificationCode(),
		ExpiredAt: time.Now().Add(s.activationCodeTTL),
	}
	usr := &domain.User{
		ID:               id,
		Name:             name,
		Surname:          surname,
		Phone:            phone,
		Email:            email,
		Roles:            roles,
		Password:         string(passwordHash),
		Activated:        activated,
		VerificationCode: newVerificationCode,
	}

	return newVerificationCode.Code, s.repo.Update(ctx, usr)
}
func (s *UserService) Delete(ctx context.Context, userID, email string) error {
	id, err := s.tokenManager.HexToObjectID(userID)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id, email)
}

func (s *UserService) Activate(ctx context.Context, userID string, activate bool) error {
	id, err := s.tokenManager.HexToObjectID(userID)
	if err != nil {
		return err
	}
	err = s.repo.Activate(ctx, id, activate)
	if err != nil {
		return err
	}
	if activate {
		err = s.repo.AddRole(ctx, id, domain.ActivatedRole)
		if err != nil {
			return err
		}
	} else {
		err = s.repo.RemoveRole(ctx, id, domain.ActivatedRole)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UserService) GenerateVerificationCode() string {
	rand.NewSource(time.Now().UnixNano())
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}

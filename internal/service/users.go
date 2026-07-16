package service

import (
	"authentication-service/internal/domain"
	"authentication-service/internal/repository"
	"authentication-service/pkg/hash"
	authManager "authentication-service/pkg/manager"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
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

func (s *UserService) SignUp(ctx context.Context, name, surname, phone, email, password string, roles []string) (string, string, error) {
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return "", "", err
	}
	code, err := s.GenerateVerificationCode()
	if err != nil {
		return "", "", err
	}
	newVerificationCode := domain.VerificationCode{
		Code:      code,
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
		return "", "", err
	}
	return user.ID, user.VerificationCode.Code, nil
}

func (s *UserService) SignIn(ctx context.Context, email string, password string) (string, []string, bool, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, false, err
	}
	ok, err := s.hasher.Matches(password, []byte(user.Password))
	if err != nil {
		return "", nil, false, err
	}
	if !ok {
		return "", nil, false, domain.ErrWrongPassword
	}
	return user.ID, user.Roles, user.Activated, nil
}

func (s *UserService) IsAdmin(ctx context.Context, userID string) (bool, error) {
	user, err := s.repo.GetByID(ctx, userID)
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
	return s.repo.GetByID(ctx, userID)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *UserService) Update(ctx context.Context, userID, name, surname, phone, email, password string, roles []string, activated bool) (string, error) {
	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return "", err
	}
	code, err := s.GenerateVerificationCode()
	if err != nil {
		return "", err
	}
	newVerificationCode := domain.VerificationCode{
		Code:      code,
		ExpiredAt: time.Now().Add(s.activationCodeTTL),
	}
	usr := &domain.User{
		ID:               userID,
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

// RefreshVerificationCode issues a new verification code for the user without
// touching any other field. It exists so callers can renew an expired code
// without routing through Update, which re-hashes the password and would
// corrupt the stored hash when handed the already-hashed value.
func (s *UserService) RefreshVerificationCode(ctx context.Context, userID string) (string, error) {
	code, err := s.GenerateVerificationCode()
	if err != nil {
		return "", err
	}
	vc := domain.VerificationCode{
		Code:      code,
		ExpiredAt: time.Now().Add(s.activationCodeTTL),
	}
	if err := s.repo.UpdateVerificationCode(ctx, userID, vc); err != nil {
		return "", err
	}
	return code, nil
}

func (s *UserService) Delete(ctx context.Context, userID, email string) error {
	return s.repo.Delete(ctx, userID, email)
}

func (s *UserService) Activate(ctx context.Context, userID string, activate bool) error {
	if err := s.repo.Activate(ctx, userID, activate); err != nil {
		return err
	}
	if activate {
		return s.repo.AddRole(ctx, userID, domain.ActivatedRole)
	}
	return s.repo.RemoveRole(ctx, userID, domain.ActivatedRole)
}

// GenerateVerificationCode returns a cryptographically random six-digit code.
func (s *UserService) GenerateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate verification code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

package service

import (
	"context"
	"errors"
	"time"

	"authentication-service/internal/domain"
	"authentication-service/internal/repository"
	"authentication-service/pkg/hash"
	"authentication-service/pkg/manager"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UsersService struct {
	repo            repository.Users
	hasher          hash.PasswordHasher
	tokenManager    auth.TokenManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	domain string
}

func NewUsersService(repo repository.Users, hasher hash.PasswordHasher, tokenManager auth.TokenManager, accessTTL, refreshTTL time.Duration, domain string) *UsersService {
	return &UsersService{
		repo:            repo,
		hasher:          hasher,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		domain:          domain,
	}
}

func (s *UsersService) SignUp(ctx context.Context, input UserSignUpInput) error {
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return err
	}

	user := domain.User{
		Email:    input.Email,
		Password: passwordHash,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return err
		}

		return err
	}
	return nil
}

func (s *UsersService) SignIn(ctx context.Context, input UserSignInInput) (Tokens, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return Tokens{}, err
		}

		return Tokens{}, err
	}

	return s.createSession(ctx, user.ID)
}

func (s *UsersService) RefreshTokens(ctx context.Context, refreshToken string) (Tokens, error) {
	student, err := s.repo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return Tokens{}, err
	}

	return s.createSession(ctx, student.ID)
}

func (s *UsersService) createSession(ctx context.Context, userId primitive.ObjectID) (Tokens, error) {
	var (
		res Tokens
		err error
	)

	res.AccessToken, err = s.tokenManager.NewJWT(userId.Hex(), s.accessTokenTTL)
	if err != nil {
		return res, err
	}

	res.RefreshToken, err = s.tokenManager.NewRefreshToken()
	if err != nil {
		return res, err
	}

	session := domain.Session{
		RefreshToken: res.RefreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTokenTTL),
	}

	err = s.repo.SetSession(ctx, userId, session)

	return res, err
}

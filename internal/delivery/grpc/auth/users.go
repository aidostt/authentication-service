package auth

import (
	"authentication-service/internal/domain"
	"authentication-service/pkg/logger"
	"context"
	"errors"
	"github.com/aidostt/protos/gen/go/reservista"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) SignUp(ctx context.Context, input *reservista.RegisterRequest) (*reservista.TokenResponse, error) {
	if input.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if input.Surname == "" {
		return nil, status.Error(codes.InvalidArgument, "surname is required")
	}
	if input.Phone == "" {
		return nil, status.Error(codes.InvalidArgument, "phone is required")
	}
	if input.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if input.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	id, err := h.services.Users.SignUp(ctx, input.GetName(), input.GetSurname(), input.GetPhone(), input.GetEmail(), input.GetPassword())
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, domain.ErrUserAlreadyExists.Error())
		}
		logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to sign up")
	}
	tokens, err := h.services.Sessions.CreateSession(ctx, id)
	if err != nil {
		logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to create session")
	}
	return &reservista.TokenResponse{Jwt: tokens.AccessToken, Rt: tokens.RefreshToken}, nil
}
func (h *Handler) SignIn(ctx context.Context, input *reservista.SignInRequest) (*reservista.TokenResponse, error) {
	if input.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if input.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	id, err := h.services.Users.SignIn(ctx, input.GetEmail(), input.GetPassword())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.AlreadyExists, domain.ErrUserNotFound.Error())
		}
		logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to sign in")
	}
	tokens, err := h.services.Sessions.CreateSession(ctx, id)
	if err != nil {
		logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to create session")
	}
	return &reservista.TokenResponse{Jwt: tokens.AccessToken, Rt: tokens.RefreshToken}, nil
}

func (h *Handler) IsAdmin(context.Context, *reservista.IsAdminRequest) (*reservista.IsAdminResponse, error) {
	//TODO: implement IsAdmin
	return nil, nil
}
func (h *Handler) SignOut(ctx context.Context, tokens *reservista.TokenRequest) (*reservista.SignOutResponse, error) {
	//TODO: implement signOut
	return nil, nil
}

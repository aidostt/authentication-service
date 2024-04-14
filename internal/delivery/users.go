package delivery

import (
	"authentication-service/internal/domain"
	"authentication-service/pkg/logger"
	"context"
	"errors"
	"github.com/aidostt/protos/gen/go/reservista/authentication"
	proto_user "github.com/aidostt/protos/gen/go/reservista/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) SignUp(ctx context.Context, input *proto_auth.RegisterRequest) (*proto_auth.TokenResponse, error) {
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
	return &proto_auth.TokenResponse{Jwt: tokens.AccessToken, Rt: tokens.RefreshToken}, nil
}
func (h *Handler) SignIn(ctx context.Context, input *proto_auth.SignInRequest) (*proto_auth.TokenResponse, error) {
	if input.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if input.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	id, err := h.services.Users.SignIn(ctx, input.GetEmail(), input.GetPassword())
	if err != nil {
		logger.Error(err)
		switch {
		case errors.Is(err, domain.ErrWrongPassword):
			return nil, status.Error(codes.InvalidArgument, domain.ErrWrongPassword.Error())
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, domain.ErrUserNotFound.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to sign in")
		}

	}
	tokens, err := h.services.Sessions.CreateSession(ctx, id)
	if err != nil {
		logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to create session")
	}
	return &proto_auth.TokenResponse{Jwt: tokens.AccessToken, Rt: tokens.RefreshToken}, nil
}

func (h *Handler) IsAdmin(context.Context, *proto_auth.IsAdminRequest) (*proto_auth.IsAdminResponse, error) {
	//TODO: implement IsAdmin
	return nil, nil
}

func (h *Handler) GetByID(ctx context.Context, input *proto_user.GetRequest) (*proto_user.UserResponse, error) {
	if input.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	user, err := h.services.Users.GetByID(ctx, input.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.InvalidArgument, "wrong id")
		default:
			return nil, status.Error(codes.Internal, "failed to get by id")
		}
	}
	return &proto_user.UserResponse{
		Name:    user.Name,
		Surname: user.Surname,
		Phone:   user.Phone,
		Email:   user.Email,
	}, nil
}

func (h *Handler) Update(ctx context.Context, input *proto_user.UpdateRequest) (*proto_user.StatusResponse, error) {
	if input.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
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
	err := h.services.Users.Update(ctx, input.GetId(), input.GetName(), input.GetSurname(), input.GetPhone(), input.GetEmail(), input.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return &proto_user.StatusResponse{Status: false}, status.Error(codes.InvalidArgument, domain.ErrUserNotFound.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to update user")
		}
	}
	return &proto_user.StatusResponse{Status: true}, nil
}

func (h *Handler) Delete(ctx context.Context, input *proto_user.GetRequest) (*proto_user.StatusResponse, error) {
	if input.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if input.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	err := h.services.Users.Delete(ctx, input.GetUserId(), input.GetEmail())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return &proto_user.StatusResponse{Status: false}, status.Error(codes.InvalidArgument, domain.ErrUserNotFound.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to delete user")
		}
	}
	return &proto_user.StatusResponse{Status: true}, nil
}

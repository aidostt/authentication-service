package delivery

import (
	"authentication-service/internal/domain"
	"authentication-service/pkg/logger"
	"context"
	"errors"
	"github.com/aidostt/protos/gen/go/reservista/authentication"
	"github.com/aidostt/protos/gen/go/reservista/user"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) SignUp(ctx context.Context, input *proto_auth.SignUpRequest) (*proto_auth.ActivationToken, error) {
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

	roles := []string{
		domain.UserRole,
	}
	id, code, err := h.services.Users.SignUp(ctx, input.GetName(), input.GetSurname(), input.GetPhone(), input.GetEmail(), input.GetPassword(), roles)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, domain.ErrUserAlreadyExists.Error())
		}
		logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to sign up: "+err.Error())
	}
	tokens, err := h.services.Sessions.CreateSession(ctx, id.Hex(), roles, false)
	if err != nil {
		logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to create session")
	}

	return &proto_auth.ActivationToken{
		ActivationToken: code,
		Tokens: &proto_auth.TokenResponse{
			Jwt: tokens.AccessToken,
			Rt:  tokens.RefreshToken,
		},
	}, nil
}

func (h *Handler) SignIn(ctx context.Context, input *proto_auth.SignInRequest) (*proto_auth.TokenResponse, error) {
	if input.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if input.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	id, roles, activated, err := h.services.Users.SignIn(ctx, input.GetEmail(), input.GetPassword())
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
	tokens, err := h.services.Sessions.CreateSession(ctx, id.Hex(), roles, activated)
	if err != nil {
		logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to create session")
	}
	return &proto_auth.TokenResponse{Jwt: tokens.AccessToken, Rt: tokens.RefreshToken}, nil
}

func (h *Handler) IsAdmin(ctx context.Context, input *proto_auth.IsAdminRequest) (*proto_auth.IsAdminResponse, error) {
	//TODO: implement IsAdmin
	if input.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	ok, err := h.services.Users.IsAdmin(ctx, input.GetUserId())
	if err != nil {

	}
	return &proto_auth.IsAdminResponse{IsAdmin: ok}, nil
}

func (h *Handler) GetByID(ctx context.Context, input *proto_user.GetRequest) (*proto_user.UserResponse, error) {
	if input.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	user, err := h.services.Users.GetByID(ctx, input.GetUserId())
	if err != nil {
		logger.Error(err)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.InvalidArgument, "wrong id")
		default:
			return nil, status.Error(codes.Internal, "failed to get by id")
		}
	}
	return &proto_user.UserResponse{
		Name:      user.Name,
		Surname:   user.Surname,
		Phone:     user.Phone,
		Email:     user.Email,
		Activated: user.Activated,
		Roles:     user.Roles,
	}, nil
}

func (h *Handler) GetByEmail(ctx context.Context, input *proto_user.GetRequest) (*proto_user.UserResponse, error) {
	if input.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	user, err := h.services.Users.GetByEmail(ctx, input.GetEmail())
	if err != nil {
		logger.Error(err)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.InvalidArgument, "wrong email")
		default:
			return nil, status.Error(codes.Internal, "failed to get by email: "+err.Error())
		}
	}
	return &proto_user.UserResponse{
		Name:      user.Name,
		Surname:   user.Surname,
		Phone:     user.Phone,
		Email:     user.Email,
		Activated: user.Activated,
		Roles:     user.Roles,
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
	_, err := h.services.Users.Update(ctx, input.GetId(), input.GetName(), input.GetSurname(), input.GetPhone(), input.GetEmail(), input.GetPassword(), input.GetRoles(), input.GetActivated())
	if err != nil {
		logger.Error(err)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return &proto_user.StatusResponse{Status: false}, status.Error(codes.InvalidArgument, domain.ErrUserNotFound.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to update user: "+err.Error())
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
		logger.Error(err)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return &proto_user.StatusResponse{Status: false}, status.Error(codes.InvalidArgument, domain.ErrUserNotFound.Error())
		default:
			return nil, status.Error(codes.Internal, "failed to delete user: "+err.Error())
		}
	}
	return &proto_user.StatusResponse{Status: true}, nil
}

func (h *Handler) Activate(ctx context.Context, input *proto_user.ActivateRequest) (*proto_user.StatusResponse, error) {
	if input.GetUserID() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	err := h.services.Users.Activate(ctx, input.GetUserID(), input.GetActivate())
	if err != nil {
		logger.Error(err)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return &proto_user.StatusResponse{Status: false}, status.Error(codes.InvalidArgument, domain.ErrUserNotFound.Error())
		default:
			return &proto_user.StatusResponse{Status: false}, status.Error(codes.Internal, "failed to activate user: "+err.Error())
		}
	}
	return &proto_user.StatusResponse{Status: true}, nil
}

func (h *Handler) VerificationCode(ctx context.Context, input *proto_user.GetRequest) (*proto_user.VerificationCodeMessage, error) {
	if input.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	user, err := h.services.Users.GetByID(ctx, input.GetUserId())
	if err != nil {
		logger.Error(err)
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.InvalidArgument, "wrong id")
		default:
			return nil, status.Error(codes.Internal, "failed to get by id")
		}
	}
	if user.VerificationCode.ExpiredAt.Before(time.Now()) {
		var code string
		code, err = h.services.Users.Update(ctx, user.ID.Hex(), user.Name, user.Surname, user.Phone, user.Email, user.Password, user.Roles, user.Activated)
		if err != nil {
			logger.Error(err)
			switch {
			case errors.Is(err, domain.ErrUserNotFound):
				return nil, status.Error(codes.InvalidArgument, domain.ErrUserNotFound.Error())
			default:
				return nil, status.Error(codes.Internal, "failed to update user: "+err.Error())
			}
		}
		return &proto_user.VerificationCodeMessage{
			Email: user.Email,
			Code:  code,
		}, nil
	}
	return &proto_user.VerificationCodeMessage{
		Email: user.Email,
		Code:  user.VerificationCode.Code,
	}, nil
}

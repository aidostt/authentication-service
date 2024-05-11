package delivery

import (
	"authentication-service/internal/domain"
	"context"
	"errors"
	"github.com/aidostt/protos/gen/go/reservista/authentication"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) Refresh(ctx context.Context, tokens *proto_auth.TokenRequest) (*proto_auth.TokenResponse, error) {
	if tokens.Jwt == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized access")
	}
	if tokens.Rt == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized access")
	}

	session, err := h.services.Sessions.GetSession(ctx, tokens.GetRt())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.Unauthenticated, "unauthorized access")
		case errors.Is(err, domain.ErrSessionExpired):
			return nil, status.Error(codes.Unauthenticated, domain.ErrSessionExpired.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}
	if session.RefreshToken != tokens.GetRt() {
		return nil, status.Error(codes.Unauthenticated, "unauthorized access")
	}

	newTokens, err := h.services.Sessions.Refresh(ctx, session.UserID, tokens.GetJwt())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnathorized), errors.Is(err, domain.ErrUserNotFound):
			return nil, status.Error(codes.Unauthenticated, "unauthorized access")
		default:
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
	}
	return &proto_auth.TokenResponse{Jwt: newTokens.AccessToken, Rt: newTokens.RefreshToken}, nil
}

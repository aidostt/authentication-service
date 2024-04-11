package auth

import (
	"authentication-service/internal/service"
	auth "authentication-service/pkg/manager"
	"github.com/aidostt/protos/gen/go/reservista"
	"google.golang.org/grpc"
)

type Handler struct {
	reservista.UnimplementedAuthServer
	auth         service.Authentication
	services     *service.Services
	tokenManager auth.TokenManager
}

func NewAuthHandler(services *service.Services, tokenManager auth.TokenManager) *Handler {
	return &Handler{
		services:     services,
		tokenManager: tokenManager,
	}
}

func (h *Handler) RegisterServerAPI(server *grpc.Server, authentication service.Authentication) {
	reservista.RegisterAuthServer(server, &Handler{auth: authentication})
}

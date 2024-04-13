package auth

import (
	"authentication-service/internal/service"
	"github.com/aidostt/protos/gen/go/reservista"
)

type Handler struct {
	reservista.UnimplementedAuthServer
	services *service.Services
}

func NewAuthHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

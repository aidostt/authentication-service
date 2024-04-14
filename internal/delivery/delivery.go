package delivery

import (
	"authentication-service/internal/service"
	proto_auth "github.com/aidostt/protos/gen/go/reservista/authentication"
	"github.com/aidostt/protos/gen/go/reservista/user"
)

type Handler struct {
	proto_auth.UnimplementedAuthServer
	proto_user.UnimplementedUserServer
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

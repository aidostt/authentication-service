package server

import (
	"authentication-service/internal/delivery"
	auth "github.com/aidostt/protos/gen/go/reservista/authentication"
	user "github.com/aidostt/protos/gen/go/reservista/user"
)

func (s *Server) RegisterServers(h *delivery.Handler) {
	auth.RegisterAuthServer(s.GrpcServer, h)
	user.RegisterUserServer(s.GrpcServer, h)
}

package server

import (
	"authentication-service/internal/config"
	"context"
	"google.golang.org/grpc"
	"net/http"
)

type Server struct {
	grpcServer *grpc.Server
}

func NewServer(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		grpcServer: &grpc.Server{

			Addr:           "localhost:" + cfg.HTTP.Port,
			Handler:        handler,
			ReadTimeout:    cfg.HTTP.ReadTimeout,
			WriteTimeout:   cfg.HTTP.WriteTimeout,
			MaxHeaderBytes: cfg.HTTP.MaxHeaderMegabytes << 20,
		},
	}
}

func (s *Server) Run() error {
	return s.grpcServer.Serve()
}

func (s *Server) Stop(ctx context.Context) {
	s.grpcServer.GracefulStop()
}

package server

import (
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	GrpcServer *grpc.Server
}

func NewServer() *Server {
	return &Server{
		GrpcServer: grpc.NewServer(),
	}

}

// Run runs gRPC server.
func (s *Server) Run(l net.Listener) error {
	return s.GrpcServer.Serve(l)

}

// Stop stops gRPC server.
func (s *Server) Stop() {
	s.GrpcServer.GracefulStop()
}

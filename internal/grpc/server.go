package grpc

import (
	"log"
	"net"

	v1 "devhive-backend/api/v1"
	"devhive-backend/internal/config"
	"devhive-backend/internal/repo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server represents the gRPC server
type Server struct {
	grpcServer *grpc.Server
	queries    *repo.Queries
	config     *config.Config
}

// New creates a new gRPC server
func New(cfg *config.Config, queries *repo.Queries) *Server {
	grpcServer := grpc.NewServer()

	// Register services
	userServer := &UserServer{queries: queries}
	projectServer := &ProjectServer{queries: queries}
	taskServer := &TaskServer{queries: queries}

	v1.RegisterUserServiceServer(grpcServer, userServer)
	v1.RegisterProjectServiceServer(grpcServer, projectServer)
	v1.RegisterTaskServiceServer(grpcServer, taskServer)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		queries:    queries,
		config:     cfg,
	}
}

// Start starts the gRPC server
func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	log.Printf("Starting gRPC server on port %s", port)
	return s.grpcServer.Serve(lis)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	log.Println("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
}

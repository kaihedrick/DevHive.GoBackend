package grpc

import (
	"log"
	"net"

	"devhive-backend/internal/config"
	"devhive-backend/internal/repo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// SimpleServer represents a basic gRPC server
type SimpleServer struct {
	grpcServer *grpc.Server
	queries    *repo.Queries
	config     *config.Config
}

// NewSimpleServer creates a new simple gRPC server
func NewSimpleServer(cfg *config.Config, queries *repo.Queries) *SimpleServer {
	grpcServer := grpc.NewServer()

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	return &SimpleServer{
		grpcServer: grpcServer,
		queries:    queries,
		config:     cfg,
	}
}

// Start starts the gRPC server
func (s *SimpleServer) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	log.Printf("Starting gRPC server on port %s", port)
	return s.grpcServer.Serve(lis)
}

// Stop gracefully stops the gRPC server
func (s *SimpleServer) Stop() {
	log.Println("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
}

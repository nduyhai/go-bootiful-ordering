package health

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// HealthServer implements the gRPC health checking protocol
type HealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

// NewHealthServer creates a new health server
func NewHealthServer() *HealthServer {
	return &HealthServer{}
}

// Check implements the gRPC health checking protocol
func (s *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

// Watch implements the gRPC health checking protocol
func (s *HealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

// RegisterHealthServer registers the health server with the gRPC server
func RegisterHealthServer(server *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(server, NewHealthServer())
}
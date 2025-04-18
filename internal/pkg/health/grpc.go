package health

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// RegisterHealthServer registers the official gRPC health server with the gRPC server
func RegisterHealthServer(server *grpc.Server) {
	healthServer := health.NewServer()
	// Set the health status for the empty service name (overall health)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)
}

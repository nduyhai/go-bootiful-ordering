package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC interceptor that collects metrics for unary RPC calls
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Start timer
		start := time.Now()

		// Process request
		resp, err := handler(ctx, req)

		// Stop timer
		duration := time.Since(start).Seconds()

		// Get status code
		st := status.Code(err)
		statusStr := st.String()

		// Record metrics
		GRPCRequestCounter.WithLabelValues(info.FullMethod, statusStr).Inc()
		GRPCRequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

		return resp, err
	}
}

// StreamServerInterceptor returns a gRPC interceptor that collects metrics for streaming RPC calls
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Start timer
		start := time.Now()

		// Process request
		err := handler(srv, ss)

		// Stop timer
		duration := time.Since(start).Seconds()

		// Get status code
		st := status.Code(err)
		statusStr := st.String()

		// Record metrics
		GRPCRequestCounter.WithLabelValues(info.FullMethod, statusStr).Inc()
		GRPCRequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

		return err
	}
}

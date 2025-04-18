package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestCounter counts the number of HTTP requests
	RequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "The total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// RequestDuration measures the duration of HTTP requests
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "The HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// GRPCRequestCounter counts the number of gRPC requests
	GRPCRequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "The total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	// GRPCRequestDuration measures the duration of gRPC requests
	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "The gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// DatabaseQueryCounter counts the number of database queries
	DatabaseQueryCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "The total number of database queries",
		},
		[]string{"operation"},
	)

	// DatabaseQueryDuration measures the duration of database queries
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "The database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// InitMetrics initializes the metrics registry
func InitMetrics(serviceName string) {
	// Register service info metric
	promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_info",
			Help: "Information about the service",
		},
		[]string{"name", "version"},
	).WithLabelValues(serviceName, "1.0.0").Set(1)
}
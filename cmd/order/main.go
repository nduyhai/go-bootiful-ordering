package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"net"
	"net/http"
	"time"

	orderv1 "go-bootiful-ordering/gen/order/v1"
	orderConfig "go-bootiful-ordering/internal/order/config"
	orderHandler "go-bootiful-ordering/internal/order/handler"
	orderRepository "go-bootiful-ordering/internal/order/repository"
	orderService "go-bootiful-ordering/internal/order/service"
	"go-bootiful-ordering/internal/pkg/config"
	"go-bootiful-ordering/internal/pkg/health"
	"go-bootiful-ordering/internal/pkg/metrics"
	"go-bootiful-ordering/internal/pkg/tracing"
)

// Route interface defines a HTTP route handler
// This is a common interface that both product and order handlers implement
type Route interface {
	Register(*gin.RouterGroup)
	Pattern() string
}

// NewGinEngine creates a new gin.Engine with the given routes
func NewGinEngine(routes []Route, tracer opentracing.Tracer) *gin.Engine {
	r := gin.Default()

	// Add OpenTracing middleware
	r.Use(tracing.GinMiddleware(tracer))

	// Add Prometheus middleware
	r.Use(metrics.GinMiddleware())

	// Register metrics endpoint
	metrics.RegisterMetricsEndpoint(r)

	// Register health check endpoint
	health.RegisterHealthEndpoint(r)

	// Create a router group for API routes
	apiGroup := r.Group("")

	// Register all routes with the router group
	for _, route := range routes {
		route.Register(apiGroup)
	}

	return r
}

func NewHTTPServer(engine *gin.Engine, cfg *config.Config) *http.Server {
	return &http.Server{
		Addr:    ":" + cfg.Server.HTTP.Port,
		Handler: engine,
	}
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(orderServer *orderHandler.GRPCOrderServer, tracer opentracing.Tracer) *grpc.Server {
	// Chain the tracing and metrics interceptors
	chainedInterceptor := grpc.ChainUnaryInterceptor(
		tracing.UnaryServerInterceptor(tracer),
		metrics.UnaryServerInterceptor(),
	)

	server := grpc.NewServer(chainedInterceptor)
	orderv1.RegisterOrderServiceServer(server, orderServer)

	// Register health check service
	health.RegisterHealthServer(server)

	return server
}

// StartHTTPServer starts the HTTP server with graceful shutdown
func StartHTTPServer(lc fx.Lifecycle, server *http.Server, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting HTTP server on " + server.Addr)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("Failed to start HTTP server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping HTTP server")
			// Use context with timeout for graceful shutdown
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				log.Error("Failed to gracefully shutdown HTTP server", zap.Error(err))
				return err
			}

			log.Info("HTTP server stopped gracefully")
			return nil
		},
	})
}

// StartGRPCServer starts the gRPC server
func StartGRPCServer(lc fx.Lifecycle, server *grpc.Server, log *zap.Logger, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			grpcAddr := ":" + cfg.Server.GRPC.Port
			listener, err := net.Listen("tcp", grpcAddr)
			if err != nil {
				log.Error("Failed to listen for gRPC", zap.Error(err))
				return err
			}

			log.Info("Starting gRPC server on " + grpcAddr)
			go func() {
				if err := server.Serve(listener); err != nil {
					log.Error("Failed to start gRPC server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping gRPC server")
			server.GracefulStop()
			return nil
		},
	})
}

// LoadConfig loads the application configuration
func LoadConfig(log *zap.Logger) (*config.Config, error) {
	cfg, err := config.LoadServiceConfig("order")
	if err != nil {
		log.Error("Failed to load configuration", zap.Error(err))
		return nil, err
	}
	return cfg, nil
}

// InitTracer initializes the OpenTracing tracer
func InitTracer(lc fx.Lifecycle, log *zap.Logger, cfg *config.Config) opentracing.Tracer {
	// Initialize tracer with configuration from YAML
	tracer, closer, err := tracing.InitTracer(cfg.Service.Name, cfg.Jaeger.HostPort())
	if err != nil {
		log.Fatal("Failed to initialize tracer", zap.Error(err))
	}

	// Register lifecycle hooks for the tracer
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("Closing tracer")
			return closer.Close()
		},
	})

	return tracer
}

// MetricsService represents the metrics service
type MetricsService struct{}

// InitMetrics initializes the Prometheus metrics
func InitMetrics(log *zap.Logger, cfg *config.Config) *MetricsService {
	log.Info("Initializing metrics")
	metrics.InitMetrics(cfg.Service.Name)
	return &MetricsService{}
}

// NewDatabaseConfig creates a database configuration from the YAML configuration
func NewDatabaseConfig(cfg *config.Config) *orderConfig.DatabaseConfig {
	return &orderConfig.DatabaseConfig{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
	}
}

func main() {
	fx.New(
		fx.Provide(fx.Annotate(
			NewHTTPServer,
			fx.ParamTags(``, ``))),
		fx.Provide(LoadConfig), // Provide the configuration
		fx.Provide(InitTracer), // Provide the tracer
		fx.Provide(InitMetrics), // Provide metrics initialization
		fx.Provide(fx.Annotate(
			NewGinEngine,
			fx.ParamTags(`group:"routes"`, ``))),

		// Order handlers
		fx.Provide(AsRoute(orderHandler.NewCreateOrderHandler)),
		fx.Provide(AsRoute(orderHandler.NewGetOrderHandler)),
		fx.Provide(AsRoute(orderHandler.NewListOrdersHandler)),
		fx.Provide(AsRoute(orderHandler.NewUpdateOrderStatusHandler)),

		// gRPC server
		fx.Provide(orderHandler.NewGRPCOrderServer),
		fx.Provide(fx.Annotate(
			NewGRPCServer,
			fx.ParamTags(``, ``, ``, ``))),

		// Logger
		fx.Provide(zap.NewExample),

		// Database configuration and connection
		fx.Provide(NewDatabaseConfig),
		fx.Provide(orderConfig.NewGormDB),

		// Order repository
		fx.Provide(fx.Annotate(orderRepository.NewGormOrderRepository, fx.As(new(orderRepository.OrderRepository)))),

		// Order service
		fx.Provide(fx.Annotate(orderService.NewDBOrderService, fx.As(new(orderService.OrderService)))),

		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Invoke(func(*gorm.DB) {}), // Add DB to invoke to ensure it's initialized
		fx.Invoke(StartHTTPServer),   // Start the HTTP server with a graceful shutdown
		fx.Invoke(StartGRPCServer),   // Start the gRPC server
	).Run()
}

// AsRoute annotates a handler constructor to be provided as a Route
func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Route)),
		fx.ResultTags(`group:"routes"`),
	)
}

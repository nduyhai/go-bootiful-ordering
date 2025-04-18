package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"net"
	"net/http"
	"time"

	productv1 "go-bootiful-ordering/gen/product/v1"
	"go-bootiful-ordering/internal/pkg/metrics"
	"go-bootiful-ordering/internal/pkg/tracing"
	productConfig "go-bootiful-ordering/internal/product/config"
	productHandler "go-bootiful-ordering/internal/product/handler"
	productRepository "go-bootiful-ordering/internal/product/repository"
	productService "go-bootiful-ordering/internal/product/service"
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

	// Create a router group for API routes
	apiGroup := r.Group("")

	// Register all routes with the router group
	for _, route := range routes {
		route.Register(apiGroup)
	}

	return r
}

func NewHTTPServer(engine *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    ":8081", // Different port from order service
		Handler: engine,
	}
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(productServer *productHandler.GRPCProductServer, tracer opentracing.Tracer) *grpc.Server {
	// Chain the tracing and metrics interceptors
	chainedInterceptor := grpc.ChainUnaryInterceptor(
		tracing.UnaryServerInterceptor(tracer),
		metrics.UnaryServerInterceptor(),
	)

	server := grpc.NewServer(chainedInterceptor)
	productv1.RegisterProductServiceServer(server, productServer)
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
func StartGRPCServer(lc fx.Lifecycle, server *grpc.Server, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := net.Listen("tcp", ":9091") // Different port from HTTP server
			if err != nil {
				log.Error("Failed to listen for gRPC", zap.Error(err))
				return err
			}

			log.Info("Starting gRPC server on :9091")
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

// InitTracer initializes the OpenTracing tracer
func InitTracer(lc fx.Lifecycle, log *zap.Logger) opentracing.Tracer {
	// Initialize tracer with default Jaeger configuration
	tracer, closer, err := tracing.InitTracer("product-service", "jaeger:6831")
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

// InitMetrics initializes the Prometheus metrics
func InitMetrics(log *zap.Logger) {
	log.Info("Initializing metrics")
	metrics.InitMetrics("product-service")
}

func main() {
	fx.New(
		fx.Provide(NewHTTPServer),
		fx.Provide(InitTracer), // Provide the tracer
		fx.Provide(InitMetrics), // Provide metrics initialization
		fx.Provide(fx.Annotate(
			NewGinEngine,
			fx.ParamTags(`group:"routes"`, ``))),

		// Product handlers
		fx.Provide(fx.Annotate(
			productHandler.NewCreateProductHandler,
			fx.As(new(Route)),
			fx.ResultTags(`group:"routes"`),
			fx.ParamTags(``, `name:"dbProductService"`),
		)),
		fx.Provide(fx.Annotate(
			productHandler.NewGetProductHandler,
			fx.As(new(Route)),
			fx.ResultTags(`group:"routes"`),
			fx.ParamTags(``, `name:"dbProductService"`),
		)),
		fx.Provide(fx.Annotate(
			productHandler.NewListProductsHandler,
			fx.As(new(Route)),
			fx.ResultTags(`group:"routes"`),
			fx.ParamTags(``, `name:"dbProductService"`),
		)),
		fx.Provide(fx.Annotate(
			productHandler.NewUpdateProductHandler,
			fx.As(new(Route)),
			fx.ResultTags(`group:"routes"`),
			fx.ParamTags(``, `name:"dbProductService"`),
		)),
		fx.Provide(fx.Annotate(
			productHandler.NewDeleteProductHandler,
			fx.As(new(Route)),
			fx.ResultTags(`group:"routes"`),
			fx.ParamTags(``, `name:"dbProductService"`),
		)),

		// gRPC server
		fx.Provide(fx.Annotate(
			productHandler.NewGRPCProductServer,
			fx.ParamTags(``, `name:"dbProductService"`))),

		fx.Provide(fx.Annotate(
			NewGRPCServer,
			fx.ParamTags(``, ``))),

		// Logger
		fx.Provide(zap.NewExample),

		// Database configuration and connection
		fx.Provide(productConfig.NewDefaultDatabaseConfig),
		fx.Provide(productConfig.NewGormDB),

		// Redis configuration and connection
		fx.Provide(productConfig.NewDefaultRedisConfig),
		fx.Provide(productConfig.NewRedisClient),

		// Product repository
		fx.Provide(productRepository.NewGormProductRepository),
		fx.Provide(fx.Annotate(
			func(redis *redis.Client, gormRepo *productRepository.GormProductRepository) productRepository.ProductRepository {
				return productRepository.NewRedisProductRepository(redis, gormRepo)
			},
			fx.As(new(productRepository.ProductRepository)),
		)),

		// Product services
		fx.Provide(fx.Annotate(
			productService.NewDBProductService,
			fx.As(new(productService.ProductService)),
			fx.ResultTags(`name:"dbProductService"`),
		)),

		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Invoke(func(*gorm.DB) {}), // Add DB to invoke to ensure it's initialized
		fx.Invoke(StartHTTPServer),   // Start the HTTP server with graceful shutdown
		fx.Invoke(StartGRPCServer),   // Start the gRPC server
	).Run()
}

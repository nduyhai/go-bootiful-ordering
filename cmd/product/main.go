package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grafana/pyroscope-go"
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
	"go-bootiful-ordering/internal/pkg/config"
	"go-bootiful-ordering/internal/pkg/health"
	"go-bootiful-ordering/internal/pkg/metrics"
	"go-bootiful-ordering/internal/pkg/migrate"
	"go-bootiful-ordering/internal/pkg/profiling"
	"go-bootiful-ordering/internal/pkg/tracing"
	productConfig "go-bootiful-ordering/internal/product/config" // Still needed for RedisConfig
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
func NewGRPCServer(productServer *productHandler.GRPCProductServer, tracer opentracing.Tracer) *grpc.Server {
	// Chain the tracing and metrics interceptors
	chainedInterceptor := grpc.ChainUnaryInterceptor(
		tracing.UnaryServerInterceptor(tracer),
		metrics.UnaryServerInterceptor(),
	)

	server := grpc.NewServer(chainedInterceptor)
	productv1.RegisterProductServiceServer(server, productServer)

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
	cfg, err := config.LoadServiceConfig("product")
	if err != nil {
		log.Error("Failed to load configuration", zap.Error(err))
		return nil, err
	}
	return cfg, nil
}

// InitTracer initializes the OpenTracing tracer
func InitTracer(lc fx.Lifecycle, log *zap.Logger, cfg *config.Config) opentracing.Tracer {
	// Initialize tracer with configuration from YAML
	// Try to use Tempo config first, fall back to Jaeger config for backward compatibility
	hostPort := cfg.Tempo.HostPort()
	if hostPort == ":" { // If Tempo config is not set
		hostPort = cfg.Jaeger.HostPort() // Use Jaeger config (which is actually TempoConfig now)
		log.Info("Using Jaeger configuration for tracing (pointing to Tempo)")
	} else {
		log.Info("Using Tempo configuration for tracing")
	}

	tracer, closer, err := tracing.InitTracer(cfg.Service.Name, hostPort)
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

// ProfilingService represents the profiling service
type ProfilingService struct {
	Profiler *pyroscope.Profiler
}

// InitProfiling initializes the Pyroscope profiler
func InitProfiling(lc fx.Lifecycle, log *zap.Logger, cfg *config.Config) (*ProfilingService, error) {
	log.Info("Initializing Pyroscope profiler")

	// Get the Pyroscope server address from config
	serverAddress := cfg.Pyroscope.ServerAddress()

	// Initialize the profiler
	profiler, err := profiling.InitProfiling(cfg.Service.Name, serverAddress)
	if err != nil {
		log.Error("Failed to initialize Pyroscope profiler", zap.Error(err))
		return nil, err
	}

	// Register lifecycle hooks for the profiler
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping Pyroscope profiler")
			profiler.Stop()
			return nil
		},
	})

	return &ProfilingService{Profiler: profiler}, nil
}

// GetDBConfig returns the database configuration from the YAML configuration
func GetDBConfig(cfg *config.Config) *config.DBConfig {
	// If DB config is not set in the YAML, create a default one for the product service
	if cfg.DB.Host == "" {
		return config.NewDefaultDBConfig("products")
	}
	return &cfg.DB
}

// NewRedisConfig creates a Redis configuration from the YAML configuration
func NewRedisConfig(cfg *config.Config) *productConfig.RedisConfig {
	return &productConfig.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
}

// RunMigrations runs database migrations
func RunMigrations(log *zap.Logger, dbConfig *config.DBConfig) error {
	log.Info("Running database migrations for product service")

	// Build DSN
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name, dbConfig.SSLMode,
	)

	// Create migration config
	migrationCfg := migrate.NewDefaultConfig("product", dsn)

	// Run migrations
	if err := migrate.Run(migrationCfg); err != nil {
		log.Error("Failed to run migrations", zap.Error(err))
		return err
	}

	log.Info("Database migrations completed successfully")
	return nil
}

func main() {
	fx.New(
		fx.Provide(fx.Annotate(
			NewHTTPServer,
			fx.ParamTags(``, ``))),
		fx.Provide(LoadConfig),    // Provide the configuration
		fx.Provide(InitTracer),    // Provide the tracer
		fx.Provide(InitMetrics),   // Provide metrics initialization
		fx.Provide(InitProfiling), // Provide profiling initialization
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
			fx.ParamTags(``, ``, ``, ``))),

		// Logger
		fx.Provide(func() *zap.Logger {
			return zap.NewExample()
		}),
		fx.Provide(func(log *zap.Logger) *zap.SugaredLogger {
			return log.Sugar()
		}),

		// Database configuration and connection
		fx.Provide(GetDBConfig),
		fx.Provide(config.NewGormDB),

		// Redis configuration and connection
		fx.Provide(NewRedisConfig),
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
		fx.Invoke(func(*gorm.DB) {}),                  // Add DB to invoke to ensure it's initialized
		fx.Invoke(func(tracer opentracing.Tracer) {}), // Add Tracer to invoke to ensure it's initialized
		fx.Invoke(func(*MetricsService) {}),           // Add MetricsService to invoke to ensure it's initialized
		fx.Invoke(func(*ProfilingService) {}),         // Add ProfilingService to invoke to ensure it's initialized
		fx.Invoke(RunMigrations),                      // Run database migrations
		fx.Invoke(StartHTTPServer),                    // Start the HTTP server with a graceful shutdown
		fx.Invoke(StartGRPCServer),                    // Start the gRPC server
	).Run()
}

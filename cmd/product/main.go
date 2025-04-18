package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"net"
	"net/http"

	productv1 "go-bootiful-ordering/gen/product/v1"
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
func NewGinEngine(routes []Route) *gin.Engine {
	r := gin.Default()

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
func NewGRPCServer(productServer *productHandler.GRPCProductServer) *grpc.Server {
	server := grpc.NewServer()
	productv1.RegisterProductServiceServer(server, productServer)
	return server
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

func main() {
	fx.New(
		fx.Provide(NewHTTPServer),
		fx.Provide(fx.Annotate(
			NewGinEngine,
			fx.ParamTags(`group:"routes"`))),

		// Product handlers
		fx.Provide(AsRoute(productHandler.NewCreateProductHandler)),
		fx.Provide(AsRoute(productHandler.NewGetProductHandler)),
		fx.Provide(AsRoute(productHandler.NewListProductsHandler)),
		fx.Provide(AsRoute(productHandler.NewUpdateProductHandler)),
		fx.Provide(AsRoute(productHandler.NewDeleteProductHandler)),

		// gRPC server
		fx.Provide(productHandler.NewGRPCProductServer),
		fx.Provide(NewGRPCServer),

		// Logger
		fx.Provide(zap.NewExample),

		// Database configuration and connection
		fx.Provide(productConfig.NewDefaultDatabaseConfig),
		fx.Provide(productConfig.NewGormDB),

		// Product repository
		fx.Provide(AsProductRepository(productRepository.NewGormProductRepository)),

		// Product services
		fx.Provide(AsProductService(productService.NewDefaultProductService)),
		fx.Provide(AsProductService(productService.NewDBProductService)),
		fx.Provide(fx.Annotate(
			productService.NewProductFactory,
			fx.ParamTags(`group:"products"`))),

		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Invoke(func(*http.Server, *gorm.DB) {}), // Add DB to invoke to ensure it's initialized
		fx.Invoke(StartGRPCServer), // Start the gRPC server
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

func AsProductService(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(productService.ProductService)),
		fx.ResultTags(`group:"products"`),
	)
}

func AsProductRepository(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(productRepository.ProductRepository)),
		fx.ResultTags(`group:"productRepositories"`),
	)
}

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

	orderv1 "go-bootiful-ordering/gen/order/v1"
	orderConfig "go-bootiful-ordering/internal/order/config"
	orderHandler "go-bootiful-ordering/internal/order/handler"
	orderRepository "go-bootiful-ordering/internal/order/repository"
	orderService "go-bootiful-ordering/internal/order/service"
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
		Addr:    ":8080", // You can make this configurable
		Handler: engine,
	}
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(orderServer *orderHandler.GRPCOrderServer) *grpc.Server {
	server := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(server, orderServer)
	return server
}

// StartGRPCServer starts the gRPC server
func StartGRPCServer(lc fx.Lifecycle, server *grpc.Server, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := net.Listen("tcp", ":9090") // Different port from HTTP server and product gRPC server
			if err != nil {
				log.Error("Failed to listen for gRPC", zap.Error(err))
				return err
			}

			log.Info("Starting gRPC server on :9090")
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

		// Order handlers
		fx.Provide(AsRoute(orderHandler.NewCreateOrderHandler)),
		fx.Provide(AsRoute(orderHandler.NewGetOrderHandler)),
		fx.Provide(AsRoute(orderHandler.NewListOrdersHandler)),
		fx.Provide(AsRoute(orderHandler.NewUpdateOrderStatusHandler)),

		// gRPC server
		fx.Provide(orderHandler.NewGRPCOrderServer),
		fx.Provide(NewGRPCServer),

		// Logger
		fx.Provide(zap.NewExample),

		// Database configuration and connection
		fx.Provide(orderConfig.NewDefaultDatabaseConfig),
		fx.Provide(orderConfig.NewGormDB),

		// Order repository
		fx.Provide(AsOrderRepository(orderRepository.NewGormOrderRepository)),

		// Order services
		fx.Provide(AsOrderService(orderService.NewDefaultOrderService)),
		fx.Provide(AsOrderService(orderService.NewRemoteOrderService)),
		fx.Provide(AsOrderService(orderService.NewDBOrderService)), // Add DB-backed service
		fx.Provide(fx.Annotate(
			orderService.NewOrderFactory,
			fx.ParamTags(`group:"orders"`))),

		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Invoke(func(*http.Server, *gorm.DB) {}), // Add DB to invoke to ensure it's initialized
		fx.Invoke(StartGRPCServer),                 // Start the gRPC server
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

func AsOrderService(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(orderService.OrderService)),
		fx.ResultTags(`group:"orderServices"`),
	)
}

func AsOrderRepository(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(orderRepository.OrderRepository)),
		fx.ResultTags(`group:"orderRepositories"`),
	)
}

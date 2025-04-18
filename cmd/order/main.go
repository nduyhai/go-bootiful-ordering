package main

import (
	"github.com/gin-gonic/gin"
	orderConfig "go-bootiful-ordering/internal/order/config"
	orderHandler "go-bootiful-ordering/internal/order/handler"
	orderRepository "go-bootiful-ordering/internal/order/repository"
	orderService "go-bootiful-ordering/internal/order/service"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
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

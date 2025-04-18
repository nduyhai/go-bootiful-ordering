package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"

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

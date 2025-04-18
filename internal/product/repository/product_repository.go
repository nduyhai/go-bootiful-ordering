package repository

import (
	"context"
	"go-bootiful-ordering/internal/product/domain"
)

// ProductRepository defines the interface for product persistence operations
type ProductRepository interface {
	// CreateProduct persists a new product and returns the created product
	CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error)
	
	// GetProduct retrieves a product by ID
	GetProduct(ctx context.Context, productID string) (*domain.Product, error)
	
	// ListProducts retrieves a list of products with pagination
	ListProducts(ctx context.Context, category string, pageSize int32, pageToken string) ([]*domain.Product, string, error)
	
	// UpdateProduct updates a product
	UpdateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error)
	
	// DeleteProduct deletes a product by ID
	DeleteProduct(ctx context.Context, productID string) error
}
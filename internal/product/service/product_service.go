package service

import (
	"context"
	"go-bootiful-ordering/internal/product/domain"
)

// ProductService defines the interface for product operations
type ProductService interface {
	CreateProduct(ctx context.Context, name, description string, price int64, stock int32, category string) (*domain.Product, error)
	GetProduct(ctx context.Context, productID string) (*domain.Product, error)
	ListProducts(ctx context.Context, category string, pageSize int32, pageToken string) ([]*domain.Product, string, error)
	UpdateProduct(ctx context.Context, productID, name, description string, price int64, stock int32, category string) (*domain.Product, error)
	DeleteProduct(ctx context.Context, productID string) error
}

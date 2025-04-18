package service

import (
	"context"
	"errors"
	"go-bootiful-ordering/internal/product/domain"
)

// ProductFactory provides a way to use multiple ProductService implementations
type ProductFactory struct {
	productSvc []ProductService
}

// NewProductFactory creates a new ProductFactory
func NewProductFactory(productSvc []ProductService) *ProductFactory {
	return &ProductFactory{productSvc: productSvc}
}

// CreateProduct tries to create a product using available services
func (f *ProductFactory) CreateProduct(ctx context.Context, name, description string, price int64, stock int32, category string) (*domain.Product, error) {
	for _, svc := range f.productSvc {
		product, err := svc.CreateProduct(ctx, name, description, price, stock, category)
		if err == nil {
			return product, nil
		}
	}
	return nil, errors.New("failed to create product")
}

// GetProduct tries to get a product using available services
func (f *ProductFactory) GetProduct(ctx context.Context, productID string) (*domain.Product, error) {
	for _, svc := range f.productSvc {
		product, err := svc.GetProduct(ctx, productID)
		if err == nil {
			return product, nil
		}
	}
	return nil, errors.New("product not found")
}

// ListProducts tries to list products using available services
func (f *ProductFactory) ListProducts(ctx context.Context, category string, pageSize int32, pageToken string) ([]*domain.Product, string, error) {
	for _, svc := range f.productSvc {
		products, nextPageToken, err := svc.ListProducts(ctx, category, pageSize, pageToken)
		if err == nil {
			return products, nextPageToken, nil
		}
	}
	return nil, "", errors.New("failed to list products")
}

// UpdateProduct tries to update a product using available services
func (f *ProductFactory) UpdateProduct(ctx context.Context, productID, name, description string, price int64, stock int32, category string) (*domain.Product, error) {
	for _, svc := range f.productSvc {
		product, err := svc.UpdateProduct(ctx, productID, name, description, price, stock, category)
		if err == nil {
			return product, nil
		}
	}
	return nil, errors.New("failed to update product")
}

// DeleteProduct tries to delete a product using available services
func (f *ProductFactory) DeleteProduct(ctx context.Context, productID string) error {
	for _, svc := range f.productSvc {
		err := svc.DeleteProduct(ctx, productID)
		if err == nil {
			return nil
		}
	}
	return errors.New("failed to delete product")
}
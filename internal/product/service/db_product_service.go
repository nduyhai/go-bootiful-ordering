package service

import (
	"context"
	"go-bootiful-ordering/internal/product/domain"
	"go-bootiful-ordering/internal/product/repository"
	"go.uber.org/zap"
)

// DBProductService provides an implementation of ProductService that uses a database repository
type DBProductService struct {
	log  *zap.SugaredLogger
	repo repository.ProductRepository
}

// NewDBProductService creates a new DBProductService
func NewDBProductService(log *zap.SugaredLogger, repo repository.ProductRepository) *DBProductService {
	return &DBProductService{
		log:  log,
		repo: repo,
	}
}

// CreateProduct creates a new product using the repository
func (s *DBProductService) CreateProduct(ctx context.Context, name, description string, price int64, stock int32, category string) (*domain.Product, error) {
	s.log.Infof("DBProductService_CreateProduct name=%s category=%s",
		name, category)

	// Create a new product domain object
	product := &domain.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Category:    category,
		Status:      domain.ProductStatusActive,
	}

	// Use the repository to persist the product
	return s.repo.CreateProduct(ctx, product)
}

// GetProduct retrieves a product by ID using the repository
func (s *DBProductService) GetProduct(ctx context.Context, productID string) (*domain.Product, error) {
	s.log.Infof("DBProductService_GetProduct productID=%s", productID)

	// Use the repository to retrieve the product
	return s.repo.GetProduct(ctx, productID)
}

// ListProducts retrieves a list of products using the repository
func (s *DBProductService) ListProducts(ctx context.Context, category string, pageSize int32, pageToken string) ([]*domain.Product, string, error) {
	s.log.Infof("DBProductService_ListProducts category=%s pageSize=%d pageToken=%s",
		category, pageSize, pageToken)

	// Use the repository to list products
	return s.repo.ListProducts(ctx, category, pageSize, pageToken)
}

// UpdateProduct updates a product using the repository
func (s *DBProductService) UpdateProduct(ctx context.Context, productID, name, description string, price int64, stock int32, category string) (*domain.Product, error) {
	s.log.Infof("DBProductService_UpdateProduct productID=%s name=%s category=%s",
		productID, name, category)

	// First, get the existing product
	existingProduct, err := s.repo.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Update the product fields
	existingProduct.Name = name
	existingProduct.Description = description
	existingProduct.Price = price
	existingProduct.Stock = stock
	existingProduct.Category = category

	// Use the repository to update the product
	return s.repo.UpdateProduct(ctx, existingProduct)
}

// DeleteProduct deletes a product using the repository
func (s *DBProductService) DeleteProduct(ctx context.Context, productID string) error {
	s.log.Infof("DBProductService_DeleteProduct productID=%s", productID)

	// Use the repository to delete the product
	return s.repo.DeleteProduct(ctx, productID)
}

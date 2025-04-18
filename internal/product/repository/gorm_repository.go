package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go-bootiful-ordering/internal/product/domain"
	"gorm.io/gorm"
	"time"
)

// GormProductRepository implements ProductRepository using GORM
type GormProductRepository struct {
	db *gorm.DB
}

// NewGormProductRepository creates a new GormProductRepository
func NewGormProductRepository(db *gorm.DB) *GormProductRepository {
	return &GormProductRepository{
		db: db,
	}
}

// CreateProduct persists a new product and returns the created product
func (r *GormProductRepository) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	// Generate a new UUID if not provided
	if product.ID == "" {
		product.ID = uuid.New().String()
	}
	
	// Set timestamps
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now
	
	// Set default status if not set
	if product.Status == domain.ProductStatusUnspecified {
		product.Status = domain.ProductStatusActive
	}
	
	// Convert domain model to database model
	productModel := FromProductDomain(product)
	
	// Begin transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	
	// Create product
	if err := tx.Create(productModel).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	
	// Return the created product
	return productModel.ToProductDomain(), nil
}

// GetProduct retrieves a product by ID
func (r *GormProductRepository) GetProduct(ctx context.Context, productID string) (*domain.Product, error) {
	var productModel ProductModel
	
	// Query product
	if err := r.db.WithContext(ctx).First(&productModel, "id = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	
	// Convert to domain model
	return productModel.ToProductDomain(), nil
}

// ListProducts retrieves a list of products with pagination
func (r *GormProductRepository) ListProducts(ctx context.Context, category string, pageSize int32, pageToken string) ([]*domain.Product, string, error) {
	var productModels []ProductModel
	
	// Build query
	query := r.db.WithContext(ctx)
	
	// Filter by category if provided
	if category != "" {
		query = query.Where("category = ?", category)
	}
	
	// Apply pagination
	if pageToken != "" {
		query = query.Where("id > ?", pageToken)
	}
	
	// Apply limit
	if pageSize > 0 {
		query = query.Limit(int(pageSize + 1)) // Fetch one extra to determine if there are more results
	}
	
	// Execute query
	if err := query.Order("id").Find(&productModels).Error; err != nil {
		return nil, "", err
	}
	
	// Determine if there are more results
	var nextPageToken string
	if len(productModels) > int(pageSize) {
		nextPageToken = productModels[len(productModels)-1].ID
		productModels = productModels[:len(productModels)-1]
	}
	
	// Convert to domain models
	products := make([]*domain.Product, len(productModels))
	for i, model := range productModels {
		products[i] = model.ToProductDomain()
	}
	
	return products, nextPageToken, nil
}

// UpdateProduct updates a product
func (r *GormProductRepository) UpdateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	// Set updated timestamp
	product.UpdatedAt = time.Now()
	
	// Convert domain model to database model
	productModel := FromProductDomain(product)
	
	// Begin transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	
	// Check if product exists
	var count int64
	if err := tx.Model(&ProductModel{}).Where("id = ?", product.ID).Count(&count).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	if count == 0 {
		tx.Rollback()
		return nil, errors.New("product not found")
	}
	
	// Update product
	if err := tx.Save(productModel).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	
	// Return the updated product
	return productModel.ToProductDomain(), nil
}

// DeleteProduct deletes a product by ID
func (r *GormProductRepository) DeleteProduct(ctx context.Context, productID string) error {
	// Begin transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	
	// Check if product exists
	var count int64
	if err := tx.Model(&ProductModel{}).Where("id = ?", productID).Count(&count).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	if count == 0 {
		tx.Rollback()
		return errors.New("product not found")
	}
	
	// Delete product
	if err := tx.Delete(&ProductModel{}, "id = ?", productID).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}
	
	return nil
}
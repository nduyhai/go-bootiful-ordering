package repository

import (
	"go-bootiful-ordering/internal/product/domain"
	"gorm.io/gorm"
	"time"
)

// ProductModel represents the database model for a product
type ProductModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string
	Description string
	Price       int64
	Stock       int32
	Category    string
	Status      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName specifies the table name for ProductModel
func (ProductModel) TableName() string {
	return "products"
}

// ToProductDomain converts a ProductModel to a domain.Product
func (m *ProductModel) ToProductDomain() *domain.Product {
	return &domain.Product{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Price:       m.Price,
		Stock:       m.Stock,
		Category:    m.Category,
		Status:      domain.ProductStatus(m.Status),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// FromProductDomain creates a ProductModel from a domain.Product
func FromProductDomain(product *domain.Product) *ProductModel {
	return &ProductModel{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		Category:    product.Category,
		Status:      int(product.Status),
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

// AutoMigrate creates or updates the database schema for product models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&ProductModel{})
}

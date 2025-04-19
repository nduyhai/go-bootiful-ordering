package domain

import (
	"time"
)

// ProductStatus represents the possible states of a product
type ProductStatus int

const (
	ProductStatusUnspecified ProductStatus = iota
	ProductStatusActive
	ProductStatusInactive
	ProductStatusOutOfStock
)

// Product represents a product in the system
type Product struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Price       int64         `json:"price"`
	Stock       int32         `json:"stock"`
	Category    string        `json:"category"`
	Status      ProductStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

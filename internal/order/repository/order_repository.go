package repository

import (
	"context"
	"go-bootiful-ordering/internal/order/domain"
	"gorm.io/gorm"
)

// OrderRepository defines the interface for order persistence operations
type OrderRepository interface {
	// CreateOrder persists a new order and returns the created order
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)

	// CreateOrderWithTx persists a new order within an existing transaction and returns the created order
	CreateOrderWithTx(ctx context.Context, tx *gorm.DB, order *domain.Order) (*domain.Order, error)

	// GetOrder retrieves an order by ID
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)

	// ListOrders retrieves a list of orders for a customer with pagination
	ListOrders(ctx context.Context, customerID string, pageSize int32, pageToken string) ([]*domain.Order, string, error)

	// UpdateOrderStatus updates the status of an order
	UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*domain.Order, error)

	// UpdateOrderStatusWithTx updates the status of an order within an existing transaction
	UpdateOrderStatusWithTx(ctx context.Context, tx *gorm.DB, orderID string, status domain.OrderStatus) (*domain.Order, error)

	// BeginTransaction starts a new transaction
	BeginTransaction(ctx context.Context) (*gorm.DB, error)
}

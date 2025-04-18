package repository

import (
	"context"
	"go-bootiful-ordering/internal/order/domain"
)

// OrderRepository defines the interface for order persistence operations
type OrderRepository interface {
	// CreateOrder persists a new order and returns the created order
	CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error)

	// GetOrder retrieves an order by ID
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)

	// ListOrders retrieves a list of orders for a customer with pagination
	ListOrders(ctx context.Context, customerID string, pageSize int32, pageToken string) ([]*domain.Order, string, error)

	// UpdateOrderStatus updates the status of an order
	UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*domain.Order, error)
}

package service

import (
	"context"
	"go-bootiful-ordering/internal/order/domain"
)

// OrderService defines the interface for order operations
type OrderService interface {
	CreateOrder(ctx context.Context, customerID string, items []domain.OrderItem) (*domain.Order, error)
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)
	ListOrders(ctx context.Context, customerID string, pageSize int32, pageToken string) ([]*domain.Order, string, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*domain.Order, error)
}

package service

import (
	"context"
	"errors"
	"go-bootiful-ordering/internal/order/domain"
)

// OrderFactory provides a way to use multiple OrderService implementations
type OrderFactory struct {
	orderSvc []OrderService
}

// NewOrderFactory creates a new OrderFactory
func NewOrderFactory(orderSvc []OrderService) *OrderFactory {
	return &OrderFactory{orderSvc: orderSvc}
}

// CreateOrder tries to create an order using available services
func (f *OrderFactory) CreateOrder(ctx context.Context, customerID string, items []domain.OrderItem) (*domain.Order, error) {
	for _, svc := range f.orderSvc {
		order, err := svc.CreateOrder(ctx, customerID, items)
		if err == nil {
			return order, nil
		}
	}
	return nil, errors.New("failed to create order")
}

// GetOrder tries to get an order using available services
func (f *OrderFactory) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	for _, svc := range f.orderSvc {
		order, err := svc.GetOrder(ctx, orderID)
		if err == nil {
			return order, nil
		}
	}
	return nil, errors.New("order not found")
}

// ListOrders tries to list orders using available services
func (f *OrderFactory) ListOrders(ctx context.Context, customerID string, pageSize int32, pageToken string) ([]*domain.Order, string, error) {
	for _, svc := range f.orderSvc {
		orders, nextPageToken, err := svc.ListOrders(ctx, customerID, pageSize, pageToken)
		if err == nil {
			return orders, nextPageToken, nil
		}
	}
	return nil, "", errors.New("failed to list orders")
}

// UpdateOrderStatus tries to update an order status using available services
func (f *OrderFactory) UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*domain.Order, error) {
	for _, svc := range f.orderSvc {
		order, err := svc.UpdateOrderStatus(ctx, orderID, status)
		if err == nil {
			return order, nil
		}
	}
	return nil, errors.New("failed to update order status")
}
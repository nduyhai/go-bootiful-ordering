package service

import (
	"context"
	"go-bootiful-ordering/internal/order/domain"
	"go-bootiful-ordering/internal/order/repository"
	"go.uber.org/zap"
)

// DBOrderService provides an implementation of OrderService that uses a database repository
type DBOrderService struct {
	log  *zap.Logger
	repo repository.OrderRepository
}

// NewDBOrderService creates a new DBOrderService
func NewDBOrderService(log *zap.Logger, repo repository.OrderRepository) *DBOrderService {
	return &DBOrderService{
		log:  log,
		repo: repo,
	}
}

// CreateOrder creates a new order using the repository
func (s *DBOrderService) CreateOrder(ctx context.Context, customerID string, items []domain.OrderItem) (*domain.Order, error) {
	s.log.Info("DBOrderService_CreateOrder", zap.String("customerID", customerID))
	
	// Create a new order domain object
	order := &domain.Order{
		CustomerID: customerID,
		Items:      items,
		Status:     domain.OrderStatusPending,
	}
	
	// Use the repository to persist the order
	return s.repo.CreateOrder(ctx, order)
}

// GetOrder retrieves an order by ID using the repository
func (s *DBOrderService) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	s.log.Info("DBOrderService_GetOrder", zap.String("orderID", orderID))
	
	// Use the repository to retrieve the order
	return s.repo.GetOrder(ctx, orderID)
}

// ListOrders retrieves a list of orders using the repository
func (s *DBOrderService) ListOrders(ctx context.Context, customerID string, pageSize int32, pageToken string) ([]*domain.Order, string, error) {
	s.log.Info("DBOrderService_ListOrders", 
		zap.String("customerID", customerID),
		zap.Int32("pageSize", pageSize),
		zap.String("pageToken", pageToken))
	
	// Use the repository to list orders
	return s.repo.ListOrders(ctx, customerID, pageSize, pageToken)
}

// UpdateOrderStatus updates the status of an order using the repository
func (s *DBOrderService) UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*domain.Order, error) {
	s.log.Info("DBOrderService_UpdateOrderStatus", 
		zap.String("orderID", orderID),
		zap.Int("status", int(status)))
	
	// Use the repository to update the order status
	return s.repo.UpdateOrderStatus(ctx, orderID, status)
}
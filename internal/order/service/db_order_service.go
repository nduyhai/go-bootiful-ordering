package service

import (
	"context"
	"go-bootiful-ordering/internal/order/domain"
	"go-bootiful-ordering/internal/order/repository"
	"go.uber.org/zap"
)

// DBOrderService provides an implementation of OrderService that uses a database repository
type DBOrderService struct {
	log        *zap.SugaredLogger
	repo       repository.OrderRepository
	outboxRepo repository.OutboxRepository
}

// NewDBOrderService creates a new DBOrderService
func NewDBOrderService(log *zap.SugaredLogger, repo repository.OrderRepository, outboxRepo repository.OutboxRepository) *DBOrderService {
	return &DBOrderService{
		log:        log,
		repo:       repo,
		outboxRepo: outboxRepo,
	}
}

// CreateOrder creates a new order using the repository
func (s *DBOrderService) CreateOrder(ctx context.Context, customerID string, items []domain.OrderItem) (*domain.Order, error) {
	s.log.Infof("DBOrderService_CreateOrder customerID=%s", customerID)

	// Create a new order domain object
	order := &domain.Order{
		CustomerID: customerID,
		Items:      items,
		Status:     domain.OrderStatusPending,
	}

	// Begin transaction
	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.log.Errorf("Failed to begin transaction: %v", err)
		return nil, err
	}

	// Create order within transaction
	createdOrder, err := s.repo.CreateOrderWithTx(ctx, tx, order)
	if err != nil {
		tx.Rollback()
		s.log.Errorf("Failed to create order: %v", err)
		return nil, err
	}

	// Create outbox entry for order created event
	outboxEntry, err := repository.NewOrderCreatedOutboxEntry(createdOrder)
	if err != nil {
		tx.Rollback()
		s.log.Errorf("Failed to create outbox entry: %v", err)
		return nil, err
	}

	// Save outbox entry within transaction
	if err := s.outboxRepo.SaveOutboxEntryWithTx(ctx, tx, outboxEntry); err != nil {
		tx.Rollback()
		s.log.Errorf("Failed to save outbox entry: %v", err)
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.log.Errorf("Failed to commit transaction: %v", err)
		return nil, err
	}

	return createdOrder, nil
}

// GetOrder retrieves an order by ID using the repository
func (s *DBOrderService) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	s.log.Infof("DBOrderService_GetOrder orderID=%s", orderID)

	// Use the repository to retrieve the order
	return s.repo.GetOrder(ctx, orderID)
}

// ListOrders retrieves a list of orders using the repository
func (s *DBOrderService) ListOrders(ctx context.Context, customerID string, pageSize int32, pageToken string) ([]*domain.Order, string, error) {
	s.log.Infof("DBOrderService_ListOrders customerID=%s pageSize=%d pageToken=%s",
		customerID, pageSize, pageToken)

	// Use the repository to list orders
	return s.repo.ListOrders(ctx, customerID, pageSize, pageToken)
}

// UpdateOrderStatus updates the status of an order using the repository
func (s *DBOrderService) UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*domain.Order, error) {
	s.log.Infof("DBOrderService_UpdateOrderStatus orderID=%s status=%d",
		orderID, int(status))

	// Begin transaction
	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		s.log.Errorf("Failed to begin transaction: %v", err)
		return nil, err
	}

	// Update order status within transaction
	updatedOrder, err := s.repo.UpdateOrderStatusWithTx(ctx, tx, orderID, status)
	if err != nil {
		tx.Rollback()
		s.log.Errorf("Failed to update order status: %v", err)
		return nil, err
	}

	// Create outbox entry for order status updated event
	outboxEntry, err := repository.NewOrderStatusUpdatedOutboxEntry(updatedOrder)
	if err != nil {
		tx.Rollback()
		s.log.Errorf("Failed to create outbox entry: %v", err)
		return nil, err
	}

	// Save outbox entry within transaction
	if err := s.outboxRepo.SaveOutboxEntryWithTx(ctx, tx, outboxEntry); err != nil {
		tx.Rollback()
		s.log.Errorf("Failed to save outbox entry: %v", err)
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.log.Errorf("Failed to commit transaction: %v", err)
		return nil, err
	}

	return updatedOrder, nil
}

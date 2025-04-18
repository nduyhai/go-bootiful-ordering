package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go-bootiful-ordering/internal/order/domain"
	"gorm.io/gorm"
	"time"
)

// GormOrderRepository implements OrderRepository using GORM
type GormOrderRepository struct {
	db *gorm.DB
}

// NewGormOrderRepository creates a new GormOrderRepository
func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
	return &GormOrderRepository{
		db: db,
	}
}

// CreateOrder persists a new order and returns the created order
func (r *GormOrderRepository) CreateOrder(ctx context.Context, order *domain.Order) (*domain.Order, error) {
	// Generate a new UUID if not provided
	if order.ID == "" {
		order.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	// Calculate total amount if not set
	if order.TotalAmount == 0 {
		for _, item := range order.Items {
			order.TotalAmount += item.Price * int64(item.Quantity)
		}
	}

	// Set default status if not set
	if order.Status == domain.OrderStatusUnspecified {
		order.Status = domain.OrderStatusPending
	}

	// Convert domain model to database model
	orderModel := FromOrderDomain(order)

	// Begin transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Create order
	if err := tx.Create(orderModel).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Return the created order
	return orderModel.ToOrderDomain(), nil
}

// GetOrder retrieves an order by ID
func (r *GormOrderRepository) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	var orderModel OrderModel

	// Query order with items
	if err := r.db.WithContext(ctx).Preload("Items").First(&orderModel, "id = ?", orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	// Convert to domain model
	return orderModel.ToOrderDomain(), nil
}

// ListOrders retrieves a list of orders for a customer with pagination
func (r *GormOrderRepository) ListOrders(ctx context.Context, customerID string, pageSize int32, pageToken string) ([]*domain.Order, string, error) {
	var orderModels []OrderModel

	// Build query
	query := r.db.WithContext(ctx).Preload("Items").Where("customer_id = ?", customerID)

	// Apply pagination
	if pageToken != "" {
		query = query.Where("id > ?", pageToken)
	}

	// Apply limit
	if pageSize > 0 {
		query = query.Limit(int(pageSize + 1)) // Fetch one extra to determine if there are more results
	}

	// Execute query
	if err := query.Order("id").Find(&orderModels).Error; err != nil {
		return nil, "", err
	}

	// Determine if there are more results
	var nextPageToken string
	if len(orderModels) > int(pageSize) {
		nextPageToken = orderModels[len(orderModels)-1].ID
		orderModels = orderModels[:len(orderModels)-1]
	}

	// Convert to domain models
	orders := make([]*domain.Order, len(orderModels))
	for i, model := range orderModels {
		orders[i] = model.ToOrderDomain()
	}

	return orders, nextPageToken, nil
}

// UpdateOrderStatus updates the status of an order
func (r *GormOrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*domain.Order, error) {
	// Begin transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Update order status
	if err := tx.Model(&OrderModel{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"status":     int(status),
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Check if order exists
	var count int64
	if err := tx.Model(&OrderModel{}).Where("id = ?", orderID).Count(&count).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if count == 0 {
		tx.Rollback()
		return nil, errors.New("order not found")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Get updated order
	return r.GetOrder(ctx, orderID)
}

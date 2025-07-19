package repository

import (
	"go-bootiful-ordering/internal/order/domain"
	"gorm.io/gorm"
	"time"
)

// OrderModel represents the database model for an order
type OrderModel struct {
	ID          string `gorm:"primaryKey"`
	CustomerID  string
	Status      int
	TotalAmount int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Items       []OrderItemModel `gorm:"foreignKey:OrderID"`
}

// OrderItemModel represents the database model for an order item
type OrderItemModel struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	OrderID   string `gorm:"index"`
	ProductID string
	Quantity  int32
	Price     int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName specifies the table name for OrderModel
func (OrderModel) TableName() string {
	return "orders"
}

// TableName specifies the table name for OrderItemModel
func (OrderItemModel) TableName() string {
	return "order_items"
}

// ToOrderDomain converts an OrderModel to a domain.Order
func (m *OrderModel) ToOrderDomain() *domain.Order {
	items := make([]domain.OrderItem, len(m.Items))
	for i, item := range m.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &domain.Order{
		ID:          m.ID,
		CustomerID:  m.CustomerID,
		Items:       items,
		Status:      domain.OrderStatus(m.Status),
		TotalAmount: m.TotalAmount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// FromOrderDomain creates an OrderModel from a domain.Order
func FromOrderDomain(order *domain.Order) *OrderModel {
	items := make([]OrderItemModel, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemModel{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &OrderModel{
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		Status:      int(order.Status),
		TotalAmount: order.TotalAmount,
		Items:       items,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}

// AutoMigrate creates or updates the database schema for order models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&OrderModel{}, &OrderItemModel{}, &OutboxModel{})
}

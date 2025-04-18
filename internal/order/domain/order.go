package domain

import (
	"time"
)

// OrderStatus represents the possible states of an order
type OrderStatus int

const (
	OrderStatusUnspecified OrderStatus = iota
	OrderStatusPending
	OrderStatusProcessing
	OrderStatusShipped
	OrderStatusDelivered
	OrderStatusCancelled
)

// OrderItem represents an item within an order
type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
	Price     int64  `json:"price"`
}

// Order represents an order in the system
type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	Status      OrderStatus `json:"status"`
	TotalAmount int64       `json:"total_amount"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

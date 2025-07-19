package repository

import (
	"encoding/json"
	"github.com/google/uuid"
	"go-bootiful-ordering/internal/order/domain"
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// EventTypeOrderCreated represents an order created event
	EventTypeOrderCreated EventType = "order_created"
	// EventTypeOrderStatusUpdated represents an order status updated event
	EventTypeOrderStatusUpdated EventType = "order_status_updated"
)

// AggregateType represents the type of aggregate
type AggregateType string

const (
	// AggregateTypeOrder represents an order aggregate
	AggregateTypeOrder AggregateType = "order"
)

// OutboxModel represents the database model for an outbox entry
type OutboxModel struct {
	ID            string    `gorm:"primaryKey;type:uuid"`
	AggregateType string    `gorm:"not null"`
	AggregateID   string    `gorm:"not null;index"`
	EventType     string    `gorm:"not null"`
	Payload       []byte    `gorm:"type:jsonb;not null"`
	CreatedAt     time.Time `gorm:"not null;index;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for OutboxModel
func (OutboxModel) TableName() string {
	return "order_outbox"
}

// NewOrderCreatedOutboxEntry creates a new outbox entry for an order created event
func NewOrderCreatedOutboxEntry(order *domain.Order) (*OutboxModel, error) {
	payload, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	return &OutboxModel{
		ID:            uuid.New().String(),
		AggregateType: string(AggregateTypeOrder),
		AggregateID:   order.ID,
		EventType:     string(EventTypeOrderCreated),
		Payload:       payload,
		CreatedAt:     time.Now(),
	}, nil
}

// NewOrderStatusUpdatedOutboxEntry creates a new outbox entry for an order status updated event
func NewOrderStatusUpdatedOutboxEntry(order *domain.Order) (*OutboxModel, error) {
	payload, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	return &OutboxModel{
		ID:            uuid.New().String(),
		AggregateType: string(AggregateTypeOrder),
		AggregateID:   order.ID,
		EventType:     string(EventTypeOrderStatusUpdated),
		Payload:       payload,
		CreatedAt:     time.Now(),
	}, nil
}

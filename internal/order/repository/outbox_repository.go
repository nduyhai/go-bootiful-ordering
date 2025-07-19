package repository

import (
	"context"
	"gorm.io/gorm"
)

// OutboxRepository defines the interface for outbox persistence operations
type OutboxRepository interface {
	// SaveOutboxEntry persists a new outbox entry
	SaveOutboxEntry(ctx context.Context, entry *OutboxModel) error

	// SaveOutboxEntryWithTx persists a new outbox entry within an existing transaction
	SaveOutboxEntryWithTx(ctx context.Context, tx *gorm.DB, entry *OutboxModel) error
}

// GormOutboxRepository implements OutboxRepository using GORM
type GormOutboxRepository struct {
	db *gorm.DB
}

// NewGormOutboxRepository creates a new GormOutboxRepository
func NewGormOutboxRepository(db *gorm.DB) *GormOutboxRepository {
	return &GormOutboxRepository{
		db: db,
	}
}

// SaveOutboxEntry persists a new outbox entry
func (r *GormOutboxRepository) SaveOutboxEntry(ctx context.Context, entry *OutboxModel) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// SaveOutboxEntryWithTx persists a new outbox entry within an existing transaction
func (r *GormOutboxRepository) SaveOutboxEntryWithTx(ctx context.Context, tx *gorm.DB, entry *OutboxModel) error {
	return tx.WithContext(ctx).Create(entry).Error
}

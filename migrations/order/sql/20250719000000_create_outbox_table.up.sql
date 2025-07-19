CREATE TABLE IF NOT EXISTS order_outbox (
    id UUID PRIMARY KEY,
    aggregate_type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_outbox_aggregate_id ON order_outbox(aggregate_id);
CREATE INDEX IF NOT EXISTS idx_order_outbox_created_at ON order_outbox(created_at);
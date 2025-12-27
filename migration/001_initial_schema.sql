-- Node Template table
CREATE TABLE IF NOT EXISTS node_template (
    id VARCHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_node_template_created_at ON node_template(created_at DESC);

-- Outbox table for reliable event publishing
CREATE TABLE IF NOT EXISTS outbox (
    id VARCHAR(26) PRIMARY KEY,
    aggregate_id VARCHAR(26) NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    retry_count INTEGER NOT NULL DEFAULT 0,

    CONSTRAINT check_retry_count CHECK (retry_count >= 0)
);

-- Index for efficient polling of unprocessed messages
CREATE INDEX IF NOT EXISTS idx_outbox_unprocessed ON outbox(created_at ASC)
    WHERE processed_at IS NULL AND retry_count < 5;

-- Index for cleanup of processed messages
CREATE INDEX IF NOT EXISTS idx_outbox_processed ON outbox(processed_at)
    WHERE processed_at IS NOT NULL;

package outbound

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"use-open-workflow.io/engine/internal/port/outbound"
)

type OutboxPostgresReadRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxPostgresReadRepository(pool *pgxpool.Pool) *OutboxPostgresReadRepository {
	return &OutboxPostgresReadRepository{pool: pool}
}

func (r *OutboxPostgresReadRepository) FindUnprocessed(ctx context.Context, limit int) ([]*outbound.OutboxMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, aggregate_id, aggregate_type, event_type, payload, created_at, retry_count
		FROM outbox
		WHERE processed_at IS NULL AND retry_count < 5
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbox: %w", err)
	}
	defer rows.Close()

	var messages []*outbound.OutboxMessage
	for rows.Next() {
		msg := &outbound.OutboxMessage{}
		if err := rows.Scan(
			&msg.ID,
			&msg.AggregateID,
			&msg.AggregateType,
			&msg.EventType,
			&msg.Payload,
			&msg.CreatedAt,
			&msg.RetryCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outbox message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return messages, nil
}

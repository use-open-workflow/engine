package outbound

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxPostgresWriteRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxPostgresWriteRepository(pool *pgxpool.Pool) *OutboxPostgresWriteRepository {
	return &OutboxPostgresWriteRepository{pool: pool}
}

func (r *OutboxPostgresWriteRepository) MarkProcessed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox
		SET processed_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to mark message as processed: %w", err)
	}
	return nil
}

func (r *OutboxPostgresWriteRepository) IncrementRetry(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox
		SET retry_count = retry_count + 1
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}
	return nil
}

func (r *OutboxPostgresWriteRepository) DeleteProcessed(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := r.pool.Exec(ctx, `
		DELETE FROM outbox
		WHERE processed_at IS NOT NULL AND processed_at < $1
	`, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete processed messages: %w", err)
	}
	return nil
}

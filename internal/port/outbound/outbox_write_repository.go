package outbound

import (
	"context"
	"time"
)

type OutboxWriteRepository interface {
	MarkProcessed(ctx context.Context, id string) error
	IncrementRetry(ctx context.Context, id string) error
	DeleteProcessed(ctx context.Context, olderThan time.Duration) error
}

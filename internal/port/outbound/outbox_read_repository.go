package outbound

import "context"

type OutboxReadRepository interface {
	FindUnprocessed(ctx context.Context, limit int) ([]*OutboxMessage, error)
}

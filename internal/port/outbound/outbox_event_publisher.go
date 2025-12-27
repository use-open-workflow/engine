package outbound

import "context"

type OutboxEventPublisher interface {
	Publish(ctx context.Context, message *OutboxMessage) error
}

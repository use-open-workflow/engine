package outbound

import "context"

type OutboxProcessor interface {
	Start(ctx context.Context) error
	Stop() error
}

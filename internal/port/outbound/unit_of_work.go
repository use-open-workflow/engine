package outbound

import "context"

type UnitOfWork interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	RegisterNew(aggregate any)
	RegisterDirty(aggregate any)
	RegisterDeleted(aggregate any)
}

type UnitOfWorkFactory interface {
	Create() UnitOfWork
}

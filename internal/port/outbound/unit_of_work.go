package outbound

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier abstracts query execution for both pool and transaction
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type UnitOfWork interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	RegisterNew(aggregate any)
	RegisterDirty(aggregate any)
	RegisterDeleted(aggregate any)

	// Querier returns the querier (tx if in transaction, or pool otherwise)
	Querier(ctx context.Context) Querier
}

type UnitOfWorkFactory interface {
	Create() UnitOfWork
}

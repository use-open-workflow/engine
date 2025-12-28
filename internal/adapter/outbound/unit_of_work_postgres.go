package outbound

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"use-open-workflow.io/engine/internal/port/outbound"
	"use-open-workflow.io/engine/pkg/domain"
)

type ctxKey string

const txKey ctxKey = "postgres_tx"

type UnitOfWorkPostgres struct {
	pool     *pgxpool.Pool
	newItems []domain.Aggregate
	dirty    []domain.Aggregate
	deleted  []domain.Aggregate
}

func NewUnitOfWorkPostgres(pool *pgxpool.Pool) *UnitOfWorkPostgres {
	return &UnitOfWorkPostgres{
		pool:     pool,
		newItems: make([]domain.Aggregate, 0),
		dirty:    make([]domain.Aggregate, 0),
		deleted:  make([]domain.Aggregate, 0),
	}
}

func (u *UnitOfWorkPostgres) Begin(ctx context.Context) (context.Context, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", err)
	}

	u.newItems = make([]domain.Aggregate, 0)
	u.dirty = make([]domain.Aggregate, 0)
	u.deleted = make([]domain.Aggregate, 0)

	return context.WithValue(ctx, txKey, tx), nil
}

func (u *UnitOfWorkPostgres) Commit(ctx context.Context) error {
	tx, ok := u.GetTx(ctx)
	if !ok {
		return fmt.Errorf("no transaction in context")
	}

	if err := u.persistOutboxEvents(ctx, tx); err != nil {
		return fmt.Errorf("failed to persist outbox events: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	u.clearAggregateEvents()

	return nil
}

func (u *UnitOfWorkPostgres) Rollback(ctx context.Context) error {
	tx, ok := u.GetTx(ctx)
	if !ok {
		return fmt.Errorf("no transaction in context")
	}

	if err := tx.Rollback(ctx); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	return nil
}

func (u *UnitOfWorkPostgres) RegisterNew(aggregate any) {
	if es, ok := aggregate.(domain.Aggregate); ok {
		u.newItems = append(u.newItems, es)
	}
}

func (u *UnitOfWorkPostgres) RegisterDirty(aggregate any) {
	if es, ok := aggregate.(domain.Aggregate); ok {
		u.dirty = append(u.dirty, es)
	}
}

func (u *UnitOfWorkPostgres) RegisterDeleted(aggregate any) {
	if es, ok := aggregate.(domain.Aggregate); ok {
		u.deleted = append(u.deleted, es)
	}
}

func (u *UnitOfWorkPostgres) GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}

func (u *UnitOfWorkPostgres) persistOutboxEvents(ctx context.Context, tx pgx.Tx) error {
	allAggregates := append(append(u.newItems, u.dirty...), u.deleted...)

	for _, aggregate := range allAggregates {
		events := aggregate.Events()
		for _, event := range events {
			payload, err := json.Marshal(event)
			if err != nil {
				return fmt.Errorf("failed to marshal event: %w", err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO outbox (id, aggregate_id, aggregate_type, event_type, payload, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`,
				event.ID(),
				event.AggregateID(),
				event.AggregateType(),
				event.EventType(),
				payload,
				event.OccurredAt(),
			)
			if err != nil {
				return fmt.Errorf("failed to insert outbox event: %w", err)
			}
		}
	}

	return nil
}

func (u *UnitOfWorkPostgres) clearAggregateEvents() {
	allAggregates := append(append(u.newItems, u.dirty...), u.deleted...)
	for _, aggregate := range allAggregates {
		aggregate.ClearEvents()
	}
}

func (u *UnitOfWorkPostgres) Querier(ctx context.Context) outbound.Querier {
	if tx, ok := u.GetTx(ctx); ok {
		return &pgxQuerier{tx: tx}
	}
	return &pgxPoolQuerier{pool: u.pool}
}

type pgxQuerier struct {
	tx pgx.Tx
}

func (q *pgxQuerier) Query(ctx context.Context, sql string, args ...any) (outbound.Rows, error) {
	rows, err := q.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRows{rows: rows}, nil
}

func (q *pgxQuerier) QueryRow(ctx context.Context, sql string, args ...any) outbound.Row {
	return q.tx.QueryRow(ctx, sql, args...)
}

func (q *pgxQuerier) Exec(ctx context.Context, sql string, args ...any) (outbound.CommandTag, error) {
	tag, err := q.tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgxCommandTag{tag: tag}, nil
}

type pgxPoolQuerier struct {
	pool *pgxpool.Pool
}

func (q *pgxPoolQuerier) Query(ctx context.Context, sql string, args ...any) (outbound.Rows, error) {
	rows, err := q.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRows{rows: rows}, nil
}

func (q *pgxPoolQuerier) QueryRow(ctx context.Context, sql string, args ...any) outbound.Row {
	return q.pool.QueryRow(ctx, sql, args...)
}

func (q *pgxPoolQuerier) Exec(ctx context.Context, sql string, args ...any) (outbound.CommandTag, error) {
	tag, err := q.pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgxCommandTag{tag: tag}, nil
}

type pgxRows struct {
	rows pgx.Rows
}

func (r *pgxRows) Close()                 { r.rows.Close() }
func (r *pgxRows) Err() error             { return r.rows.Err() }
func (r *pgxRows) Next() bool             { return r.rows.Next() }
func (r *pgxRows) Scan(dest ...any) error { return r.rows.Scan(dest...) }

// pgxCommandTag wraps pgconn.CommandTag to implement outbound.CommandTag
type pgxCommandTag struct {
	tag pgconn.CommandTag
}

func (t *pgxCommandTag) RowsAffected() int64 { return t.tag.RowsAffected() }

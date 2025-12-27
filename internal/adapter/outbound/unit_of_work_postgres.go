package outbound

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	portOutbound "use-open-workflow.io/engine/internal/port/outbound"
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

// Querier returns the transaction if one is active, otherwise returns the pool
func (u *UnitOfWorkPostgres) Querier(ctx context.Context) portOutbound.Querier {
	if tx, ok := u.GetTx(ctx); ok {
		return tx
	}
	return u.pool
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

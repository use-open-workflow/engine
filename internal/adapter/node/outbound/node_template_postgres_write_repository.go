package outbound

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"use-open-workflow.io/engine/internal/adapter/outbound"
	"use-open-workflow.io/engine/internal/domain/node/aggregate"
)

type NodeTemplatePostgresWriteRepository struct {
	pool *pgxpool.Pool
	uow  *outbound.UnitOfWorkPostgres
}

func NewNodeTemplatePostgresWriteRepository(
	pool *pgxpool.Pool,
	uow *outbound.UnitOfWorkPostgres,
) *NodeTemplatePostgresWriteRepository {
	return &NodeTemplatePostgresWriteRepository{
		pool: pool,
		uow:  uow,
	}
}

func (r *NodeTemplatePostgresWriteRepository) getQuerier(ctx context.Context) querier {
	if tx, ok := r.uow.GetTx(ctx); ok {
		return tx
	}
	return r.pool
}

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (r *NodeTemplatePostgresWriteRepository) Save(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
	q := r.getQuerier(ctx)

	_, err := q.Exec(ctx, `
		INSERT INTO node_templates (id, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
	`, nodeTemplate.ID, nodeTemplate.Name)

	if err != nil {
		return fmt.Errorf("failed to save node template: %w", err)
	}

	r.uow.RegisterNew(nodeTemplate)

	return nil
}

func (r *NodeTemplatePostgresWriteRepository) Update(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
	q := r.getQuerier(ctx)

	_, err := q.Exec(ctx, `
		UPDATE node_templates
		SET name = $1, updated_at = NOW()
		WHERE id = $2
	`, nodeTemplate.Name, nodeTemplate.ID)

	if err != nil {
		return fmt.Errorf("failed to update node template: %w", err)
	}

	r.uow.RegisterDirty(nodeTemplate)

	return nil
}

func (r *NodeTemplatePostgresWriteRepository) Delete(ctx context.Context, id string) error {
	q := r.getQuerier(ctx)

	_, err := q.Exec(ctx, `
		DELETE FROM node_templates
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete node template: %w", err)
	}

	return nil
}

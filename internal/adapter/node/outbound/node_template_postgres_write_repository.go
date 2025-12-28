package outbound

import (
	"context"
	"fmt"

	"use-open-workflow.io/engine/internal/domain/node/aggregate"
	portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

type NodeTemplatePostgresWriteRepository struct {
	uow portOutbound.UnitOfWork
}

func NewNodeTemplatePostgresWriteRepository(
	uow portOutbound.UnitOfWork,
) *NodeTemplatePostgresWriteRepository {
	return &NodeTemplatePostgresWriteRepository{
		uow: uow,
	}
}

func (r *NodeTemplatePostgresWriteRepository) Save(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
	q := r.uow.Querier(ctx)

	_, err := q.Exec(ctx, `
		INSERT INTO node_templates (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`, nodeTemplate.ID, nodeTemplate.Name, nodeTemplate.CreatedAt, nodeTemplate.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save node template: %w", err)
	}

	r.uow.RegisterNew(nodeTemplate)

	return nil
}

func (r *NodeTemplatePostgresWriteRepository) Update(ctx context.Context, nodeTemplate *aggregate.NodeTemplate) error {
	q := r.uow.Querier(ctx)

	_, err := q.Exec(ctx, `
		UPDATE node_templates
		SET name = $1, updated_at = $2
		WHERE id = $3
	`, nodeTemplate.Name, nodeTemplate.UpdatedAt, nodeTemplate.ID)

	if err != nil {
		return fmt.Errorf("failed to update node template: %w", err)
	}

	r.uow.RegisterDirty(nodeTemplate)

	return nil
}

func (r *NodeTemplatePostgresWriteRepository) Delete(ctx context.Context, id string) error {
	q := r.uow.Querier(ctx)

	_, err := q.Exec(ctx, `
		DELETE FROM node_templates
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete node template: %w", err)
	}

	return nil
}

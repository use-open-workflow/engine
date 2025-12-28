package outbound

import (
	"context"
	"fmt"
	"time"

	"use-open-workflow.io/engine/internal/domain/node/aggregate"
	portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

type NodeTemplatePostgresReadRepository struct {
	uow portOutbound.UnitOfWork
}

func NewNodeTemplatePostgresReadRepository(
	uow portOutbound.UnitOfWork,
) *NodeTemplatePostgresReadRepository {
	return &NodeTemplatePostgresReadRepository{
		uow: uow,
	}
}

func (r *NodeTemplatePostgresReadRepository) FindMany(ctx context.Context) ([]*aggregate.NodeTemplate, error) {
	q := r.uow.Querier(ctx)

	rows, err := q.Query(ctx, `
		SELECT id, name, created_at, updated_at
		FROM node_templates
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query node templates: %w", err)
	}
	defer rows.Close()

	var templates []*aggregate.NodeTemplate
	for rows.Next() {
		var id, name string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &name, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan node template: %w", err)
		}

		template := aggregate.ReconstituteNodeTemplate(id, name, createdAt, updatedAt)
		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return templates, nil
}

func (r *NodeTemplatePostgresReadRepository) FindByID(ctx context.Context, id string) (*aggregate.NodeTemplate, error) {
	q := r.uow.Querier(ctx)

	var name string
	var createdAt, updatedAt time.Time
	err := q.QueryRow(ctx, `
		SELECT id, name, created_at, updated_at
		FROM node_templates
		WHERE id = $1
	`, id).Scan(&id, &name, &createdAt, &updatedAt)

	if err != nil && err.Error() == "no rows in result set" {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query node template: %w", err)
	}

	return aggregate.ReconstituteNodeTemplate(id, name, createdAt, updatedAt), nil
}

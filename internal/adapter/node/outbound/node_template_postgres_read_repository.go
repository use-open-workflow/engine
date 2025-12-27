package outbound

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"use-open-workflow.io/engine/internal/domain/node/aggregate"
)

type NodeTemplatePostgresReadRepository struct {
	pool *pgxpool.Pool
}

func NewNodeTemplatePostgresReadRepository(pool *pgxpool.Pool) *NodeTemplatePostgresReadRepository {
	return &NodeTemplatePostgresReadRepository{
		pool: pool,
	}
}

func (r *NodeTemplatePostgresReadRepository) FindMany(ctx context.Context) ([]*aggregate.NodeTemplate, error) {
	rows, err := r.pool.Query(ctx, `
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
		var createdAt, updatedAt interface{}
		if err := rows.Scan(&id, &name, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan node template: %w", err)
		}

		template := aggregate.ReconstituteNodeTemplate(id, name)
		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return templates, nil
}

func (r *NodeTemplatePostgresReadRepository) FindByID(ctx context.Context, id string) (*aggregate.NodeTemplate, error) {
	var name string
	var createdAt, updatedAt interface{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, created_at, updated_at
		FROM node_templates
		WHERE id = $1
	`, id).Scan(&id, &name, &createdAt, &updatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query node template: %w", err)
	}

	return aggregate.ReconstituteNodeTemplate(id, name), nil
}

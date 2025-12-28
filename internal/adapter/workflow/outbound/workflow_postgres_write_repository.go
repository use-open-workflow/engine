package outbound

import (
	"context"
	"encoding/json"
	"fmt"

	"use-open-workflow.io/engine/internal/domain/workflow/aggregate"
	"use-open-workflow.io/engine/internal/domain/workflow/entity"
	portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

type WorkflowPostgresWriteRepository struct {
	uow portOutbound.UnitOfWork
}

func NewWorkflowPostgresWriteRepository(
	uow portOutbound.UnitOfWork,
) *WorkflowPostgresWriteRepository {
	return &WorkflowPostgresWriteRepository{
		uow: uow,
	}
}

func (r *WorkflowPostgresWriteRepository) Save(ctx context.Context, workflow *aggregate.Workflow) error {
	q := r.uow.Querier(ctx)

	_, err := q.Exec(ctx, `
		INSERT INTO workflow (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, workflow.ID, workflow.Name, workflow.Description, workflow.CreatedAt, workflow.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save workflow: %w", err)
	}

	r.uow.RegisterNew(workflow)
	return nil
}

func (r *WorkflowPostgresWriteRepository) Update(ctx context.Context, workflow *aggregate.Workflow) error {
	q := r.uow.Querier(ctx)

	_, err := q.Exec(ctx, `
		UPDATE workflow
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4
	`, workflow.Name, workflow.Description, workflow.UpdatedAt, workflow.ID)

	if err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	r.uow.RegisterDirty(workflow)
	return nil
}

func (r *WorkflowPostgresWriteRepository) Delete(ctx context.Context, id string) error {
	q := r.uow.Querier(ctx)

	// CASCADE will handle node_definition and edge deletion
	_, err := q.Exec(ctx, `
		DELETE FROM workflow
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	return nil
}

func (r *WorkflowPostgresWriteRepository) SaveNodeDefinition(ctx context.Context, nodeDef *entity.NodeDefinition) error {
	q := r.uow.Querier(ctx)

	configJSON, err := json.Marshal(nodeDef.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = q.Exec(ctx, `
		INSERT INTO node_definition (id, workflow_id, node_template_id, name, config, position_x, position_y, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, nodeDef.ID, nodeDef.WorkflowID, nodeDef.NodeTemplateID, nodeDef.Name, configJSON, nodeDef.PositionX, nodeDef.PositionY, nodeDef.CreatedAt, nodeDef.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save node definition: %w", err)
	}

	return nil
}

func (r *WorkflowPostgresWriteRepository) UpdateNodeDefinition(ctx context.Context, nodeDef *entity.NodeDefinition) error {
	q := r.uow.Querier(ctx)

	configJSON, err := json.Marshal(nodeDef.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = q.Exec(ctx, `
		UPDATE node_definition
		SET name = $1, config = $2, position_x = $3, position_y = $4, updated_at = $5
		WHERE id = $6
	`, nodeDef.Name, configJSON, nodeDef.PositionX, nodeDef.PositionY, nodeDef.UpdatedAt, nodeDef.ID)

	if err != nil {
		return fmt.Errorf("failed to update node definition: %w", err)
	}

	return nil
}

func (r *WorkflowPostgresWriteRepository) DeleteNodeDefinition(ctx context.Context, id string) error {
	q := r.uow.Querier(ctx)

	// CASCADE will handle edge deletion
	_, err := q.Exec(ctx, `
		DELETE FROM node_definition
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete node definition: %w", err)
	}

	return nil
}

func (r *WorkflowPostgresWriteRepository) SaveEdge(ctx context.Context, edge *entity.Edge) error {
	q := r.uow.Querier(ctx)

	_, err := q.Exec(ctx, `
		INSERT INTO edge (id, workflow_id, from_node_definition_id, to_node_definition_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, edge.ID, edge.WorkflowID, edge.FromNodeDefinitionID, edge.ToNodeDefinitionID, edge.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to save edge: %w", err)
	}

	return nil
}

func (r *WorkflowPostgresWriteRepository) DeleteEdge(ctx context.Context, id string) error {
	q := r.uow.Querier(ctx)

	_, err := q.Exec(ctx, `
		DELETE FROM edge
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete edge: %w", err)
	}

	return nil
}

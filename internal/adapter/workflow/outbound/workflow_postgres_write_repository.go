package outbound

import (
	"context"
	"fmt"

	"use-open-workflow.io/engine/internal/domain/workflow/aggregate"
	portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

// WorkflowPostgresWriteRepository implements WorkflowWriteRepository.
type WorkflowPostgresWriteRepository struct {
	uow portOutbound.UnitOfWork
}

// NewWorkflowPostgresWriteRepository creates a new repository.
func NewWorkflowPostgresWriteRepository(uow portOutbound.UnitOfWork) *WorkflowPostgresWriteRepository {
	return &WorkflowPostgresWriteRepository{uow: uow}
}

// Save persists a new workflow with all its child entities.
func (r *WorkflowPostgresWriteRepository) Save(ctx context.Context, workflow *aggregate.Workflow) error {
	q := r.uow.Querier(ctx)

	// 1. Insert workflow
	_, err := q.Exec(ctx, `
		INSERT INTO workflow (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`, workflow.ID, workflow.Name, workflow.CreatedAt, workflow.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save workflow: %w", err)
	}

	// 2. Insert node definitions
	for _, nd := range workflow.NodeDefinitions {
		_, err := q.Exec(ctx, `
			INSERT INTO node_definition (id, workflow_id, node_template_id, name, position_x, position_y)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, nd.ID, nd.WorkflowID, nd.NodeTemplateID, nd.Name, nd.PositionX, nd.PositionY)
		if err != nil {
			return fmt.Errorf("failed to save node definition: %w", err)
		}
	}

	// 3. Insert edges
	for _, edge := range workflow.Edges {
		_, err := q.Exec(ctx, `
			INSERT INTO edge (id, workflow_id, from_node_id, to_node_id)
			VALUES ($1, $2, $3, $4)
		`, edge.ID, edge.WorkflowID, edge.FromNodeID, edge.ToNodeID)
		if err != nil {
			return fmt.Errorf("failed to save edge: %w", err)
		}
	}

	// 4. Register aggregate for event publishing
	r.uow.RegisterNew(workflow)

	return nil
}

// Update updates an existing workflow and syncs all child entities.
// Uses delete-and-insert strategy for simplicity.
func (r *WorkflowPostgresWriteRepository) Update(ctx context.Context, workflow *aggregate.Workflow) error {
	q := r.uow.Querier(ctx)

	// 1. Update workflow
	_, err := q.Exec(ctx, `
		UPDATE workflow
		SET name = $1, updated_at = $2
		WHERE id = $3
	`, workflow.Name, workflow.UpdatedAt, workflow.ID)
	if err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	// 2. Delete existing edges (must delete before nodes due to FK)
	_, err = q.Exec(ctx, `DELETE FROM edge WHERE workflow_id = $1`, workflow.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing edges: %w", err)
	}

	// 3. Delete existing node definitions
	_, err = q.Exec(ctx, `DELETE FROM node_definition WHERE workflow_id = $1`, workflow.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing node definitions: %w", err)
	}

	// 4. Re-insert node definitions
	for _, nd := range workflow.NodeDefinitions {
		_, err := q.Exec(ctx, `
			INSERT INTO node_definition (id, workflow_id, node_template_id, name, position_x, position_y)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, nd.ID, nd.WorkflowID, nd.NodeTemplateID, nd.Name, nd.PositionX, nd.PositionY)
		if err != nil {
			return fmt.Errorf("failed to insert node definition: %w", err)
		}
	}

	// 5. Re-insert edges
	for _, edge := range workflow.Edges {
		_, err := q.Exec(ctx, `
			INSERT INTO edge (id, workflow_id, from_node_id, to_node_id)
			VALUES ($1, $2, $3, $4)
		`, edge.ID, edge.WorkflowID, edge.FromNodeID, edge.ToNodeID)
		if err != nil {
			return fmt.Errorf("failed to insert edge: %w", err)
		}
	}

	// 6. Register aggregate for event publishing
	r.uow.RegisterDirty(workflow)

	return nil
}

// Delete removes a workflow by ID.
// Child entities are deleted via CASCADE.
func (r *WorkflowPostgresWriteRepository) Delete(ctx context.Context, id string) error {
	q := r.uow.Querier(ctx)

	_, err := q.Exec(ctx, `DELETE FROM workflow WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	return nil
}

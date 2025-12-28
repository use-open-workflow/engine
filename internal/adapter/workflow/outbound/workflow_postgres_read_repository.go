package outbound

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"use-open-workflow.io/engine/internal/domain/workflow/aggregate"
	"use-open-workflow.io/engine/internal/domain/workflow/entity"
	portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

type WorkflowPostgresReadRepository struct {
	uow portOutbound.UnitOfWork
}

func NewWorkflowPostgresReadRepository(
	uow portOutbound.UnitOfWork,
) *WorkflowPostgresReadRepository {
	return &WorkflowPostgresReadRepository{
		uow: uow,
	}
}

func (r *WorkflowPostgresReadRepository) FindMany(ctx context.Context) ([]*aggregate.Workflow, error) {
	q := r.uow.Querier(ctx)

	rows, err := q.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM workflow
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*aggregate.Workflow
	for rows.Next() {
		var id, name, description string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &name, &description, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}

		workflow := aggregate.ReconstituteWorkflow(id, name, description, createdAt, updatedAt)
		workflows = append(workflows, workflow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return workflows, nil
}

func (r *WorkflowPostgresReadRepository) FindByID(ctx context.Context, id string) (*aggregate.Workflow, error) {
	q := r.uow.Querier(ctx)

	// 1. Fetch workflow
	var name, description string
	var createdAt, updatedAt time.Time
	err := q.QueryRow(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM workflow
		WHERE id = $1
	`, id).Scan(&id, &name, &description, &createdAt, &updatedAt)

	if err != nil && err.Error() == "no rows in result set" {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow: %w", err)
	}

	workflow := aggregate.ReconstituteWorkflow(id, name, description, createdAt, updatedAt)

	// 2. Fetch NodeDefinitions
	nodeRows, err := q.Query(ctx, `
		SELECT id, workflow_id, node_template_id, name, config, position_x, position_y, created_at, updated_at
		FROM node_definition
		WHERE workflow_id = $1
		ORDER BY created_at ASC
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query node definitions: %w", err)
	}
	defer nodeRows.Close()

	var nodeDefs []*entity.NodeDefinition
	for nodeRows.Next() {
		var ndID, wfID, ntID, ndName string
		var configJSON []byte
		var posX, posY float64
		var ndCreatedAt, ndUpdatedAt time.Time

		if err := nodeRows.Scan(&ndID, &wfID, &ntID, &ndName, &configJSON, &posX, &posY, &ndCreatedAt, &ndUpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan node definition: %w", err)
		}

		var config map[string]interface{}
		if configJSON != nil {
			if err := json.Unmarshal(configJSON, &config); err != nil {
				return nil, fmt.Errorf("failed to unmarshal config: %w", err)
			}
		}

		nodeDef := entity.ReconstituteNodeDefinition(ndID, wfID, ntID, ndName, config, posX, posY, ndCreatedAt, ndUpdatedAt)
		nodeDefs = append(nodeDefs, nodeDef)
	}
	if err := nodeRows.Err(); err != nil {
		return nil, fmt.Errorf("node definition row iteration error: %w", err)
	}
	workflow.SetNodeDefinitions(nodeDefs)

	// 3. Fetch Edges
	edgeRows, err := q.Query(ctx, `
		SELECT id, workflow_id, from_node_definition_id, to_node_definition_id, created_at
		FROM edge
		WHERE workflow_id = $1
		ORDER BY created_at ASC
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges: %w", err)
	}
	defer edgeRows.Close()

	var edges []*entity.Edge
	for edgeRows.Next() {
		var eID, eWfID, fromID, toID string
		var eCreatedAt time.Time

		if err := edgeRows.Scan(&eID, &eWfID, &fromID, &toID, &eCreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}

		edge := entity.ReconstituteEdge(eID, eWfID, fromID, toID, eCreatedAt)
		edges = append(edges, edge)
	}
	if err := edgeRows.Err(); err != nil {
		return nil, fmt.Errorf("edge row iteration error: %w", err)
	}
	workflow.SetEdges(edges)

	return workflow, nil
}

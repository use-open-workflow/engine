package outbound

import (
	"context"
	"fmt"
	"time"

	"use-open-workflow.io/engine/internal/domain/workflow/aggregate"
	portOutbound "use-open-workflow.io/engine/internal/port/outbound"
)

// WorkflowPostgresReadRepository implements WorkflowReadRepository.
type WorkflowPostgresReadRepository struct {
	uow portOutbound.UnitOfWork
}

// NewWorkflowPostgresReadRepository creates a new repository.
func NewWorkflowPostgresReadRepository(uow portOutbound.UnitOfWork) *WorkflowPostgresReadRepository {
	return &WorkflowPostgresReadRepository{uow: uow}
}

// FindMany returns all workflows with their child entities.
func (r *WorkflowPostgresReadRepository) FindMany(ctx context.Context) ([]*aggregate.Workflow, error) {
	q := r.uow.Querier(ctx)

	// 1. Query all workflows
	rows, err := q.Query(ctx, `
		SELECT id, name, created_at, updated_at
		FROM workflow
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflows: %w", err)
	}
	defer rows.Close()

	// 2. Build workflow map
	workflowMap := make(map[string]*aggregate.Workflow)
	workflowIDs := make([]string, 0)

	for rows.Next() {
		var id, name string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &name, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}
		workflow := aggregate.ReconstituteWorkflow(id, name, createdAt, updatedAt, nil, nil)
		workflowMap[id] = workflow
		workflowIDs = append(workflowIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	if len(workflowIDs) == 0 {
		return []*aggregate.Workflow{}, nil
	}

	// 3. Load all node definitions for these workflows
	if err := r.loadNodeDefinitions(ctx, workflowMap); err != nil {
		return nil, err
	}

	// 4. Load all edges for these workflows
	if err := r.loadEdges(ctx, workflowMap); err != nil {
		return nil, err
	}

	// 5. Preserve order
	result := make([]*aggregate.Workflow, len(workflowIDs))
	for i, id := range workflowIDs {
		result[i] = workflowMap[id]
	}

	return result, nil
}

// FindByID returns a workflow by ID with all child entities.
func (r *WorkflowPostgresReadRepository) FindByID(ctx context.Context, id string) (*aggregate.Workflow, error) {
	q := r.uow.Querier(ctx)

	// 1. Query workflow
	var name string
	var createdAt, updatedAt time.Time
	err := q.QueryRow(ctx, `
		SELECT id, name, created_at, updated_at
		FROM workflow
		WHERE id = $1
	`, id).Scan(&id, &name, &createdAt, &updatedAt)

	if err != nil && err.Error() == "no rows in result set" {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow: %w", err)
	}

	// 2. Query node definitions
	nodeRows, err := q.Query(ctx, `
		SELECT id, workflow_id, node_template_id, name, position_x, position_y
		FROM node_definition
		WHERE workflow_id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query node definitions: %w", err)
	}
	defer nodeRows.Close()

	nodeDefinitions := make([]*aggregate.NodeDefinition, 0)
	for nodeRows.Next() {
		var ndID, wfID, ntID, ndName string
		var posX, posY float64
		if err := nodeRows.Scan(&ndID, &wfID, &ntID, &ndName, &posX, &posY); err != nil {
			return nil, fmt.Errorf("failed to scan node definition: %w", err)
		}
		nodeDefinitions = append(nodeDefinitions, aggregate.ReconstituteNodeDefinition(ndID, wfID, ntID, ndName, posX, posY))
	}
	if err := nodeRows.Err(); err != nil {
		return nil, fmt.Errorf("node row iteration error: %w", err)
	}

	// 3. Query edges
	edgeRows, err := q.Query(ctx, `
		SELECT id, workflow_id, from_node_id, to_node_id
		FROM edge
		WHERE workflow_id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges: %w", err)
	}
	defer edgeRows.Close()

	edges := make([]*aggregate.Edge, 0)
	for edgeRows.Next() {
		var eID, wfID, fromID, toID string
		if err := edgeRows.Scan(&eID, &wfID, &fromID, &toID); err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}
		edges = append(edges, aggregate.ReconstituteEdge(eID, wfID, fromID, toID))
	}
	if err := edgeRows.Err(); err != nil {
		return nil, fmt.Errorf("edge row iteration error: %w", err)
	}

	return aggregate.ReconstituteWorkflow(id, name, createdAt, updatedAt, nodeDefinitions, edges), nil
}

// loadNodeDefinitions loads node definitions for all workflows in map.
func (r *WorkflowPostgresReadRepository) loadNodeDefinitions(ctx context.Context, workflowMap map[string]*aggregate.Workflow) error {
	q := r.uow.Querier(ctx)

	rows, err := q.Query(ctx, `
		SELECT id, workflow_id, node_template_id, name, position_x, position_y
		FROM node_definition
		WHERE workflow_id = ANY($1)
	`, r.getWorkflowIDs(workflowMap))
	if err != nil {
		return fmt.Errorf("failed to query node definitions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, workflowID, nodeTemplateID, name string
		var posX, posY float64
		if err := rows.Scan(&id, &workflowID, &nodeTemplateID, &name, &posX, &posY); err != nil {
			return fmt.Errorf("failed to scan node definition: %w", err)
		}
		nd := aggregate.ReconstituteNodeDefinition(id, workflowID, nodeTemplateID, name, posX, posY)
		if wf, ok := workflowMap[workflowID]; ok {
			wf.NodeDefinitions = append(wf.NodeDefinitions, nd)
		}
	}
	return rows.Err()
}

// loadEdges loads edges for all workflows in map.
func (r *WorkflowPostgresReadRepository) loadEdges(ctx context.Context, workflowMap map[string]*aggregate.Workflow) error {
	q := r.uow.Querier(ctx)

	rows, err := q.Query(ctx, `
		SELECT id, workflow_id, from_node_id, to_node_id
		FROM edge
		WHERE workflow_id = ANY($1)
	`, r.getWorkflowIDs(workflowMap))
	if err != nil {
		return fmt.Errorf("failed to query edges: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, workflowID, fromNodeID, toNodeID string
		if err := rows.Scan(&id, &workflowID, &fromNodeID, &toNodeID); err != nil {
			return fmt.Errorf("failed to scan edge: %w", err)
		}
		edge := aggregate.ReconstituteEdge(id, workflowID, fromNodeID, toNodeID)
		if wf, ok := workflowMap[workflowID]; ok {
			wf.Edges = append(wf.Edges, edge)
		}
	}
	return rows.Err()
}

func (r *WorkflowPostgresReadRepository) getWorkflowIDs(workflowMap map[string]*aggregate.Workflow) []string {
	ids := make([]string, 0, len(workflowMap))
	for id := range workflowMap {
		ids = append(ids, id)
	}
	return ids
}

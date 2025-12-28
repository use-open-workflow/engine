package outbound

import (
	"context"

	"use-open-workflow.io/engine/internal/domain/workflow/aggregate"
)

// WorkflowWriteRepository defines write operations for workflows.
type WorkflowWriteRepository interface {
	// Save persists a new workflow with all its child entities.
	Save(ctx context.Context, workflow *aggregate.Workflow) error

	// Update updates an existing workflow and syncs all child entities.
	// This performs a full sync: inserts new, updates existing, deletes removed.
	Update(ctx context.Context, workflow *aggregate.Workflow) error

	// Delete removes a workflow by ID (cascade deletes child entities via FK).
	Delete(ctx context.Context, id string) error
}

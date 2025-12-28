package inbound

import "use-open-workflow.io/engine/internal/domain/workflow/aggregate"

// WorkflowMapper converts between aggregates and DTOs.
type WorkflowMapper interface {
	// To converts a Workflow aggregate to a WorkflowDTO.
	To(workflow *aggregate.Workflow) (*WorkflowDTO, error)

	// ToList converts a slice of Workflow aggregates to WorkflowDTOs.
	ToList(workflows []*aggregate.Workflow) ([]*WorkflowDTO, error)
}

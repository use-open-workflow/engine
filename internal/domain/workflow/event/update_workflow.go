package event

import (
	"use-open-workflow.io/engine/pkg/domain"
	"use-open-workflow.io/engine/pkg/id"
)

// UpdateWorkflow is emitted when a workflow is updated.
type UpdateWorkflow struct {
	domain.BaseEvent
	WorkflowID string `json:"workflow_id"`
	Name       string `json:"name"`
}

// NewUpdateWorkflow creates a new UpdateWorkflow event.
func NewUpdateWorkflow(idFactory id.Factory, workflowID, name string) *UpdateWorkflow {
	return &UpdateWorkflow{
		BaseEvent: domain.NewBaseEvent(
			idFactory.New(),
			workflowID,
			"Workflow",
			"UpdateWorkflow",
		),
		WorkflowID: workflowID,
		Name:       name,
	}
}

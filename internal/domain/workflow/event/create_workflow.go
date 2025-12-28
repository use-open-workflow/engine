package event

import (
	"use-open-workflow.io/engine/pkg/domain"
	"use-open-workflow.io/engine/pkg/id"
)

// CreateWorkflow is emitted when a new workflow is created.
type CreateWorkflow struct {
	domain.BaseEvent
	WorkflowID string `json:"workflow_id"`
	Name       string `json:"name"`
}

// NewCreateWorkflow creates a new CreateWorkflow event.
func NewCreateWorkflow(idFactory id.Factory, workflowID, name string) *CreateWorkflow {
	return &CreateWorkflow{
		BaseEvent: domain.NewBaseEvent(
			idFactory.New(),
			workflowID,
			"Workflow",
			"CreateWorkflow",
		),
		WorkflowID: workflowID,
		Name:       name,
	}
}

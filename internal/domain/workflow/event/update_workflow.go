package event

import (
	"use-open-workflow.io/engine/pkg/domain"
	"use-open-workflow.io/engine/pkg/id"
)

type UpdateWorkflow struct {
	domain.BaseEvent
	WorkflowID  string `json:"workflow_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewUpdateWorkflow(idFactory id.Factory, workflowID, name, description string) *UpdateWorkflow {
	return &UpdateWorkflow{
		BaseEvent: domain.NewBaseEvent(
			idFactory.New(),
			workflowID,
			"Workflow",
			"UpdateWorkflow",
		),
		WorkflowID:  workflowID,
		Name:        name,
		Description: description,
	}
}

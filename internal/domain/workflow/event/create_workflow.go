package event

import (
	"use-open-workflow.io/engine/pkg/domain"
	"use-open-workflow.io/engine/pkg/id"
)

type CreateWorkflow struct {
	domain.BaseEvent
	WorkflowID  string `json:"workflow_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewCreateWorkflow(idFactory id.Factory, workflowID, name, description string) *CreateWorkflow {
	return &CreateWorkflow{
		BaseEvent: domain.NewBaseEvent(
			idFactory.New(),
			workflowID,
			"Workflow",
			"CreateWorkflow",
		),
		WorkflowID:  workflowID,
		Name:        name,
		Description: description,
	}
}

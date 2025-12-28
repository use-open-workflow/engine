package event

import (
	"use-open-workflow.io/engine/pkg/domain"
	"use-open-workflow.io/engine/pkg/id"
)

type RemoveNodeDefinition struct {
	domain.BaseEvent
	WorkflowID       string `json:"workflow_id"`
	NodeDefinitionID string `json:"node_definition_id"`
}

func NewRemoveNodeDefinition(
	idFactory id.Factory,
	workflowID, nodeDefinitionID string,
) *RemoveNodeDefinition {
	return &RemoveNodeDefinition{
		BaseEvent: domain.NewBaseEvent(
			idFactory.New(),
			workflowID,
			"Workflow",
			"RemoveNodeDefinition",
		),
		WorkflowID:       workflowID,
		NodeDefinitionID: nodeDefinitionID,
	}
}

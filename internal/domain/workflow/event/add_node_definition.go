package event

import (
	"use-open-workflow.io/engine/pkg/domain"
	"use-open-workflow.io/engine/pkg/id"
)

type AddNodeDefinition struct {
	domain.BaseEvent
	WorkflowID       string `json:"workflow_id"`
	NodeDefinitionID string `json:"node_definition_id"`
	NodeTemplateID   string `json:"node_template_id"`
	Name             string `json:"name"`
}

func NewAddNodeDefinition(
	idFactory id.Factory,
	workflowID, nodeDefinitionID, nodeTemplateID, name string,
) *AddNodeDefinition {
	return &AddNodeDefinition{
		BaseEvent: domain.NewBaseEvent(
			idFactory.New(),
			workflowID,
			"Workflow",
			"AddNodeDefinition",
		),
		WorkflowID:       workflowID,
		NodeDefinitionID: nodeDefinitionID,
		NodeTemplateID:   nodeTemplateID,
		Name:             name,
	}
}

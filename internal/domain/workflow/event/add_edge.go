package event

import (
	"use-open-workflow.io/engine/pkg/domain"
	"use-open-workflow.io/engine/pkg/id"
)

type AddEdge struct {
	domain.BaseEvent
	WorkflowID           string `json:"workflow_id"`
	EdgeID               string `json:"edge_id"`
	FromNodeDefinitionID string `json:"from_node_definition_id"`
	ToNodeDefinitionID   string `json:"to_node_definition_id"`
}

func NewAddEdge(
	idFactory id.Factory,
	workflowID, edgeID, fromNodeDefinitionID, toNodeDefinitionID string,
) *AddEdge {
	return &AddEdge{
		BaseEvent: domain.NewBaseEvent(
			idFactory.New(),
			workflowID,
			"Workflow",
			"AddEdge",
		),
		WorkflowID:           workflowID,
		EdgeID:               edgeID,
		FromNodeDefinitionID: fromNodeDefinitionID,
		ToNodeDefinitionID:   toNodeDefinitionID,
	}
}

package event

import (
	"use-open-workflow.io/engine/pkg/domain"
	"use-open-workflow.io/engine/pkg/id"
)

type RemoveEdge struct {
	domain.BaseEvent
	WorkflowID string `json:"workflow_id"`
	EdgeID     string `json:"edge_id"`
}

func NewRemoveEdge(idFactory id.Factory, workflowID, edgeID string) *RemoveEdge {
	return &RemoveEdge{
		BaseEvent: domain.NewBaseEvent(
			idFactory.New(),
			workflowID,
			"Workflow",
			"RemoveEdge",
		),
		WorkflowID: workflowID,
		EdgeID:     edgeID,
	}
}

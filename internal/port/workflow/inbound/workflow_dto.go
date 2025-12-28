package inbound

import "time"

// WorkflowDTO represents a Workflow for API responses.
type WorkflowDTO struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	NodeDefinitions []NodeDefinitionDTO `json:"nodeDefinitions"`
	Edges           []EdgeDTO           `json:"edges"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
}

// NodeDefinitionDTO represents a NodeDefinition for API responses.
type NodeDefinitionDTO struct {
	ID             string  `json:"id"`
	WorkflowID     string  `json:"workflowId"`
	NodeTemplateID string  `json:"nodeTemplateId"`
	Name           string  `json:"name"`
	PositionX      float64 `json:"positionX"`
	PositionY      float64 `json:"positionY"`
}

// EdgeDTO represents an Edge for API responses.
type EdgeDTO struct {
	ID         string `json:"id"`
	WorkflowID string `json:"workflowId"`
	FromNodeID string `json:"fromNodeId"`
	ToNodeID   string `json:"toNodeId"`
}

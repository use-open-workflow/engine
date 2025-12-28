package inbound

import "time"

// WorkflowDTO represents a workflow for API responses
type WorkflowDTO struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	NodeDefinitions []*NodeDefinitionDTO `json:"nodeDefinitions"`
	Edges           []*EdgeDTO           `json:"edges"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
}

// NodeDefinitionDTO represents a node definition for API responses
type NodeDefinitionDTO struct {
	ID             string                 `json:"id"`
	WorkflowID     string                 `json:"workflowId"`
	NodeTemplateID string                 `json:"nodeTemplateId"`
	Name           string                 `json:"name"`
	Config         map[string]interface{} `json:"config,omitempty"`
	PositionX      float64                `json:"positionX"`
	PositionY      float64                `json:"positionY"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

// EdgeDTO represents an edge for API responses
type EdgeDTO struct {
	ID                   string    `json:"id"`
	WorkflowID           string    `json:"workflowId"`
	FromNodeDefinitionID string    `json:"fromNodeDefinitionId"`
	ToNodeDefinitionID   string    `json:"toNodeDefinitionId"`
	CreatedAt            time.Time `json:"createdAt"`
}

// Input structs for write operations
type CreateWorkflowInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateWorkflowInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddNodeDefinitionInput struct {
	NodeTemplateID string                 `json:"nodeTemplateId"`
	Name           string                 `json:"name"`
	Config         map[string]interface{} `json:"config,omitempty"`
	PositionX      float64                `json:"positionX"`
	PositionY      float64                `json:"positionY"`
}

type UpdateNodeDefinitionInput struct {
	Name      string                 `json:"name,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
	PositionX *float64               `json:"positionX,omitempty"`
	PositionY *float64               `json:"positionY,omitempty"`
}

type AddEdgeInput struct {
	FromNodeDefinitionID string `json:"fromNodeDefinitionId"`
	ToNodeDefinitionID   string `json:"toNodeDefinitionId"`
}

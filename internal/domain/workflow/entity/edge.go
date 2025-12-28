package entity

import (
	"time"

	"use-open-workflow.io/engine/pkg/domain"
)

// Edge connects two NodeDefinitions within a Workflow
type Edge struct {
	domain.BaseEntity
	WorkflowID           string
	FromNodeDefinitionID string
	ToNodeDefinitionID   string
	CreatedAt            time.Time
}

// NewEdge creates a new Edge
func NewEdge(
	id string,
	workflowID string,
	fromNodeDefinitionID string,
	toNodeDefinitionID string,
) *Edge {
	return &Edge{
		BaseEntity:           domain.NewBaseEntity(id),
		WorkflowID:           workflowID,
		FromNodeDefinitionID: fromNodeDefinitionID,
		ToNodeDefinitionID:   toNodeDefinitionID,
		CreatedAt:            time.Now().UTC(),
	}
}

// ReconstituteEdge recreates entity from database
func ReconstituteEdge(
	id string,
	workflowID string,
	fromNodeDefinitionID string,
	toNodeDefinitionID string,
	createdAt time.Time,
) *Edge {
	return &Edge{
		BaseEntity:           domain.NewBaseEntity(id),
		WorkflowID:           workflowID,
		FromNodeDefinitionID: fromNodeDefinitionID,
		ToNodeDefinitionID:   toNodeDefinitionID,
		CreatedAt:            createdAt,
	}
}

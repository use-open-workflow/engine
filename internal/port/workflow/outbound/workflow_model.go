package outbound

import "time"

// WorkflowModel represents a workflow database row.
type WorkflowModel struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NodeDefinitionModel represents a node_definition database row.
type NodeDefinitionModel struct {
	ID             string
	WorkflowID     string
	NodeTemplateID string
	Name           string
	PositionX      float64
	PositionY      float64
}

// EdgeModel represents an edge database row.
type EdgeModel struct {
	ID         string
	WorkflowID string
	FromNodeID string
	ToNodeID   string
}

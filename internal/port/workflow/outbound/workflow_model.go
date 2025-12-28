package outbound

import "time"

// WorkflowModel represents workflow in database
type WorkflowModel struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewWorkflowModel() *WorkflowModel {
	return &WorkflowModel{}
}

// NodeDefinitionModel represents node_definition in database
type NodeDefinitionModel struct {
	ID             string
	WorkflowID     string
	NodeTemplateID string
	Name           string
	Config         map[string]interface{}
	PositionX      float64
	PositionY      float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewNodeDefinitionModel() *NodeDefinitionModel {
	return &NodeDefinitionModel{}
}

// EdgeModel represents edge in database
type EdgeModel struct {
	ID                   string
	WorkflowID           string
	FromNodeDefinitionID string
	ToNodeDefinitionID   string
	CreatedAt            time.Time
}

func NewEdgeModel() *EdgeModel {
	return &EdgeModel{}
}

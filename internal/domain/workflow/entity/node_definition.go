package entity

import (
	"time"

	"use-open-workflow.io/engine/pkg/domain"
)

// NodeDefinition represents a node instance within a workflow
type NodeDefinition struct {
	domain.BaseEntity
	WorkflowID     string
	NodeTemplateID string
	Name           string
	Config         map[string]interface{}
	PositionX      float64
	PositionY      float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewNodeDefinition creates a new NodeDefinition
func NewNodeDefinition(
	id string,
	workflowID string,
	nodeTemplateID string,
	name string,
	config map[string]interface{},
	positionX float64,
	positionY float64,
) *NodeDefinition {
	now := time.Now().UTC()
	return &NodeDefinition{
		BaseEntity:     domain.NewBaseEntity(id),
		WorkflowID:     workflowID,
		NodeTemplateID: nodeTemplateID,
		Name:           name,
		Config:         config,
		PositionX:      positionX,
		PositionY:      positionY,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// ReconstituteNodeDefinition recreates entity from database
func ReconstituteNodeDefinition(
	id string,
	workflowID string,
	nodeTemplateID string,
	name string,
	config map[string]interface{},
	positionX float64,
	positionY float64,
	createdAt time.Time,
	updatedAt time.Time,
) *NodeDefinition {
	return &NodeDefinition{
		BaseEntity:     domain.NewBaseEntity(id),
		WorkflowID:     workflowID,
		NodeTemplateID: nodeTemplateID,
		Name:           name,
		Config:         config,
		PositionX:      positionX,
		PositionY:      positionY,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
}

// UpdatePosition updates the visual position
func (n *NodeDefinition) UpdatePosition(x, y float64) {
	n.PositionX = x
	n.PositionY = y
	n.UpdatedAt = time.Now().UTC()
}

// UpdateConfig updates the node configuration
func (n *NodeDefinition) UpdateConfig(config map[string]interface{}) {
	n.Config = config
	n.UpdatedAt = time.Now().UTC()
}

// UpdateName updates the display name
func (n *NodeDefinition) UpdateName(name string) {
	n.Name = name
	n.UpdatedAt = time.Now().UTC()
}

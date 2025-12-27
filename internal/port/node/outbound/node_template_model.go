package outbound

import "time"

type NodeTemplateModel struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewNodeTemplateModel() *NodeTemplateModel {
	return &NodeTemplateModel{}
}

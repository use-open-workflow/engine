package outbound

type NodeTemplateModel struct {
	ID   string
	Name string
}

func NewNodeTemplateModel() *NodeTemplateModel {
	return &NodeTemplateModel{}
}

package inbound

import "context"

type CreateNodeTemplateInput struct {
	Name string `json:"name"`
}

type UpdateNodeTemplateInput struct {
	Name string `json:"name"`
}

type NodeTemplateWriteService interface {
	Create(ctx context.Context, input CreateNodeTemplateInput) (*NodeTemplateDTO, error)
	Update(ctx context.Context, id string, input UpdateNodeTemplateInput) (*NodeTemplateDTO, error)
	Delete(ctx context.Context, id string) error
}

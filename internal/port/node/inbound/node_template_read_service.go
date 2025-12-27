package inbound

import "context"

type NodeTemplateReadService interface {
	List(ctx context.Context) ([]*NodeTemplateDTO, error)
	GetByID(ctx context.Context, id string) (*NodeTemplateDTO, error)
}

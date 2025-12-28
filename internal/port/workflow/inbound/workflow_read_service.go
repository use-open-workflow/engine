package inbound

import "context"

type WorkflowReadService interface {
	List(ctx context.Context) ([]*WorkflowDTO, error)
	GetByID(ctx context.Context, id string) (*WorkflowDTO, error)
}

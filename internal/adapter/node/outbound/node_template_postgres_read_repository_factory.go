package outbound

import (
	nodeOutbound "use-open-workflow.io/engine/internal/port/node/outbound"
	"use-open-workflow.io/engine/internal/port/outbound"
)

type NodeTemplatePostgresReadRepositoryFactory struct{}

func NewNodeTemplatePostgresReadRepositoryFactory() *NodeTemplatePostgresReadRepositoryFactory {
	return &NodeTemplatePostgresReadRepositoryFactory{}
}

func (f *NodeTemplatePostgresReadRepositoryFactory) Create(uow outbound.UnitOfWork) nodeOutbound.NodeTemplateReadRepository {
	return NewNodeTemplatePostgresReadRepository(uow)
}

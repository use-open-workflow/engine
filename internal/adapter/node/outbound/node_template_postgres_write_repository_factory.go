package outbound

import (
	nodeOutbound "use-open-workflow.io/engine/internal/port/node/outbound"
	"use-open-workflow.io/engine/internal/port/outbound"
)

type NodeTemplatePostgresWriteRepositoryFactory struct{}

func NewNodeTemplatePostgresWriteRepositoryFactory() *NodeTemplatePostgresWriteRepositoryFactory {
	return &NodeTemplatePostgresWriteRepositoryFactory{}
}

func (f *NodeTemplatePostgresWriteRepositoryFactory) Create(uow outbound.UnitOfWork) nodeOutbound.NodeTemplateWriteRepository {
	return NewNodeTemplatePostgresWriteRepository(uow)
}

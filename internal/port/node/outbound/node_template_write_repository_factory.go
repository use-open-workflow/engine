package outbound

import "use-open-workflow.io/engine/internal/port/outbound"

type NodeTemplateWriteRepositoryFactory interface {
	Create(uow outbound.UnitOfWork) NodeTemplateWriteRepository
}

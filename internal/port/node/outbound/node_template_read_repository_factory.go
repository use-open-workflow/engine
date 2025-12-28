package outbound

import "use-open-workflow.io/engine/internal/port/outbound"

type NodeTemplateReadRepositoryFactory interface {
	Create(uow outbound.UnitOfWork) NodeTemplateReadRepository
}

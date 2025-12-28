package outbound

import "use-open-workflow.io/engine/internal/port/outbound"

type WorkflowWriteRepositoryFactory interface {
	Create(uow outbound.UnitOfWork) WorkflowWriteRepository
}

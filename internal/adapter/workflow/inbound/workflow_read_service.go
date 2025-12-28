package inbound

import (
	"context"

	"use-open-workflow.io/engine/internal/port/outbound"
	"use-open-workflow.io/engine/internal/port/workflow/inbound"
	workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
)

type WorkflowReadService struct {
	uowFactory            outbound.UnitOfWorkFactory
	readRepositoryFactory workflowOutbound.WorkflowReadRepositoryFactory
	mapper                inbound.WorkflowMapper
}

func NewWorkflowReadService(
	uowFactory outbound.UnitOfWorkFactory,
	readRepositoryFactory workflowOutbound.WorkflowReadRepositoryFactory,
	mapper inbound.WorkflowMapper,
) *WorkflowReadService {
	return &WorkflowReadService{
		uowFactory:            uowFactory,
		readRepositoryFactory: readRepositoryFactory,
		mapper:                mapper,
	}
}

func (s *WorkflowReadService) List(ctx context.Context) ([]*inbound.WorkflowDTO, error) {
	uow := s.uowFactory.Create()
	readRepo := s.readRepositoryFactory.Create(uow)

	workflows, err := readRepo.FindMany(ctx)
	if err != nil {
		return nil, err
	}

	workflowDTOs := make([]*inbound.WorkflowDTO, len(workflows))
	for i, w := range workflows {
		dto, err := s.mapper.ToWorkflowDTO(w)
		if err != nil {
			return nil, err
		}
		workflowDTOs[i] = dto
	}

	return workflowDTOs, nil
}

func (s *WorkflowReadService) GetByID(ctx context.Context, id string) (*inbound.WorkflowDTO, error) {
	uow := s.uowFactory.Create()
	readRepo := s.readRepositoryFactory.Create(uow)

	workflow, err := readRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, nil
	}

	return s.mapper.ToWorkflowDTO(workflow)
}

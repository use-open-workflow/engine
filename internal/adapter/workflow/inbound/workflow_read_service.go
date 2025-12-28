package inbound

import (
	"context"
	"fmt"

	"use-open-workflow.io/engine/internal/port/outbound"
	"use-open-workflow.io/engine/internal/port/workflow/inbound"
	workflowOutbound "use-open-workflow.io/engine/internal/port/workflow/outbound"
)

// WorkflowReadService implements the WorkflowReadService interface.
type WorkflowReadService struct {
	uowFactory            outbound.UnitOfWorkFactory
	readRepositoryFactory workflowOutbound.WorkflowReadRepositoryFactory
	mapper                inbound.WorkflowMapper
}

// NewWorkflowReadService creates a new WorkflowReadService.
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

// List returns all workflows.
func (s *WorkflowReadService) List(ctx context.Context) ([]*inbound.WorkflowDTO, error) {
	// 1. Create UoW and repository
	uow := s.uowFactory.Create()
	readRepo := s.readRepositoryFactory.Create(uow)

	// 2. Fetch workflows
	workflows, err := readRepo.FindMany(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflows: %w", err)
	}

	// 3. Convert to DTOs
	return s.mapper.ToList(workflows)
}

// GetByID returns a workflow by ID.
func (s *WorkflowReadService) GetByID(ctx context.Context, id string) (*inbound.WorkflowDTO, error) {
	// 1. Create UoW and repository
	uow := s.uowFactory.Create()
	readRepo := s.readRepositoryFactory.Create(uow)

	// 2. Fetch workflow
	workflow, err := readRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow: %w", err)
	}
	if workflow == nil {
		return nil, nil // Not found
	}

	// 3. Convert to DTO
	return s.mapper.To(workflow)
}

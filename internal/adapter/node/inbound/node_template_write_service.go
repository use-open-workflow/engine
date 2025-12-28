package inbound

import (
	"context"
	"fmt"

	"use-open-workflow.io/engine/internal/domain/node/aggregate"
	"use-open-workflow.io/engine/internal/port/node/inbound"
	nodeOutbound "use-open-workflow.io/engine/internal/port/node/outbound"
	"use-open-workflow.io/engine/internal/port/outbound"
	"use-open-workflow.io/engine/pkg/id"
)

type NodeTemplateWriteService struct {
	uowFactory             outbound.UnitOfWorkFactory
	writeRepositoryFactory nodeOutbound.NodeTemplateWriteRepositoryFactory
	readRepositoryFactory  nodeOutbound.NodeTemplateReadRepositoryFactory
	factory                *aggregate.NodeTemplateFactory
	mapper                 inbound.NodeTemplateMapper
	idFactory              id.Factory
}

func NewNodeTemplateWriteService(
	uowFactory outbound.UnitOfWorkFactory,
	writeRepositoryFactory nodeOutbound.NodeTemplateWriteRepositoryFactory,
	readRepositoryFactory nodeOutbound.NodeTemplateReadRepositoryFactory,
	factory *aggregate.NodeTemplateFactory,
	mapper inbound.NodeTemplateMapper,
	idFactory id.Factory,
) *NodeTemplateWriteService {
	return &NodeTemplateWriteService{
		uowFactory:             uowFactory,
		writeRepositoryFactory: writeRepositoryFactory,
		readRepositoryFactory:  readRepositoryFactory,
		factory:                factory,
		mapper:                 mapper,
		idFactory:              idFactory,
	}
}

func (s *NodeTemplateWriteService) Create(ctx context.Context, input inbound.CreateNodeTemplateInput) (*inbound.NodeTemplateDTO, error) {
	uow := s.uowFactory.Create()

	// Create repository bound to THIS UoW
	writeRepo := s.writeRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	nodeTemplate := s.factory.Make(input.Name)

	// Save using the UoW-bound repository
	if err = writeRepo.Save(txCtx, nodeTemplate); err != nil {
		return nil, fmt.Errorf("failed to save node template: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.To(nodeTemplate)
}

func (s *NodeTemplateWriteService) Update(ctx context.Context, id string, input inbound.UpdateNodeTemplateInput) (*inbound.NodeTemplateDTO, error) {
	uow := s.uowFactory.Create()

	// Create repositories bound to THIS UoW
	writeRepo := s.writeRepositoryFactory.Create(uow)
	readRepo := s.readRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	nodeTemplate, err := readRepo.FindByID(txCtx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find node template: %w", err)
	}
	if nodeTemplate == nil {
		return nil, fmt.Errorf("node template not found: %s", id)
	}

	// Update aggregate (this adds UpdateNodeTemplate event)
	nodeTemplate.UpdateName(s.idFactory, input.Name)

	// Update using UoW-bound repository
	if err = writeRepo.Update(txCtx, nodeTemplate); err != nil {
		return nil, fmt.Errorf("failed to update node template: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return s.mapper.To(nodeTemplate)
}

func (s *NodeTemplateWriteService) Delete(ctx context.Context, id string) error {
	uow := s.uowFactory.Create()

	// Create repository bound to THIS UoW
	writeRepo := s.writeRepositoryFactory.Create(uow)

	txCtx, err := uow.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			uow.Rollback(txCtx)
		}
	}()

	// Delete using UoW-bound repository
	if err = writeRepo.Delete(txCtx, id); err != nil {
		return fmt.Errorf("failed to delete node template: %w", err)
	}

	if err = uow.Commit(txCtx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

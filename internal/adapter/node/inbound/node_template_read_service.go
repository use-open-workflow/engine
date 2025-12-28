package inbound

import (
	"context"

	"use-open-workflow.io/engine/internal/port/node/inbound"
	nodeOutbound "use-open-workflow.io/engine/internal/port/node/outbound"
	"use-open-workflow.io/engine/internal/port/outbound"
)

type NodeTemplateReadService struct {
	uowFactory            outbound.UnitOfWorkFactory
	readRepositoryFactory nodeOutbound.NodeTemplateReadRepositoryFactory
	mapper                inbound.NodeTemplateMapper
}

func NewNodeTemplateReadService(
	uowFactory outbound.UnitOfWorkFactory,
	readRepositoryFactory nodeOutbound.NodeTemplateReadRepositoryFactory,
	mapper inbound.NodeTemplateMapper,
) *NodeTemplateReadService {
	return &NodeTemplateReadService{
		uowFactory:            uowFactory,
		readRepositoryFactory: readRepositoryFactory,
		mapper:                mapper,
	}
}

func (s *NodeTemplateReadService) List(ctx context.Context) ([]*inbound.NodeTemplateDTO, error) {
	uow := s.uowFactory.Create()
	readRepo := s.readRepositoryFactory.Create(uow)

	nodeTemplates, err := readRepo.FindMany(ctx)
	if err != nil {
		return nil, err
	}

	nodeTemplateDTOs := make([]*inbound.NodeTemplateDTO, len(nodeTemplates))
	for i, v := range nodeTemplates {
		nodeTemplateDTO, err := s.mapper.To(v)
		if err != nil {
			return nil, err
		}
		nodeTemplateDTOs[i] = nodeTemplateDTO
	}

	return nodeTemplateDTOs, nil
}

func (s *NodeTemplateReadService) GetByID(ctx context.Context, id string) (*inbound.NodeTemplateDTO, error) {
	uow := s.uowFactory.Create()
	readRepo := s.readRepositoryFactory.Create(uow)

	nodeTemplate, err := readRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if nodeTemplate == nil {
		return nil, nil
	}

	return s.mapper.To(nodeTemplate)
}

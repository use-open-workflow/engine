package di

import (
	adapterInbound "use-open-workflow.io/engine/internal/adapter/node/inbound"
	adapterOutbound "use-open-workflow.io/engine/internal/adapter/node/outbound"
	"use-open-workflow.io/engine/internal/port/node/inbound"
)

type Container struct {
	NodeTemplateReadService inbound.NodeTemplateReadService
}

func NewContainer() *Container {
	nodeTemplateInboundMapper := adapterInbound.NewNodeTemplateMapper()
	nodeTemplateOutboundMapper := adapterOutbound.NewNodeTemplateMapper()

	nodeTemplateReadRepository := adapterOutbound.NewStaticNodeTemplateReadRepository(nodeTemplateOutboundMapper)
	nodeTemplateReadService := adapterInbound.NewNodeTemplateReadService(nodeTemplateInboundMapper, nodeTemplateReadRepository)

	return &Container{
		NodeTemplateReadService: nodeTemplateReadService,
	}
}

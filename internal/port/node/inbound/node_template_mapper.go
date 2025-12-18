package inbound

import "use-open-workflow.io/engine/internal/domain/node/aggregate"

type NodeTemplateMapper interface {
	To(*aggregate.NodeTemplate) (*NodeTemplateDTO, error)
}

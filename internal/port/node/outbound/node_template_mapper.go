package outbound

import "use-open-workflow.io/engine/internal/domain/node/aggregate"

type NodeTemplateMapper interface {
	From(*NodeTemplateModel) (*aggregate.NodeTemplate, error)
	To(*aggregate.NodeTemplate) (*NodeTemplateModel, error)
}

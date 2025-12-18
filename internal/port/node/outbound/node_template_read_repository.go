package outbound

import "use-open-workflow.io/engine/internal/domain/node/aggregate"

type NodeTemplateRepository interface {
	FindMany() ([]*aggregate.NodeTemplate, error)
	FindByID(id string) (*aggregate.NodeTemplate, error)
}

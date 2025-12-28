package aggregate

import "use-open-workflow.io/engine/pkg/domain"

// Edge represents a directed connection between two NodeDefinitions.
// It is a child entity owned by the Workflow aggregate.
type Edge struct {
	domain.BaseEntity
	WorkflowID string
	FromNodeID string
	ToNodeID   string
}

// newEdge creates a new Edge entity.
// Called internally by Workflow.AddEdge()
func newEdge(id, workflowID, fromNodeID, toNodeID string) *Edge {
	return &Edge{
		BaseEntity: domain.NewBaseEntity(id),
		WorkflowID: workflowID,
		FromNodeID: fromNodeID,
		ToNodeID:   toNodeID,
	}
}

// ReconstituteEdge recreates an Edge from persistence.
// Used by repository when loading from database.
func ReconstituteEdge(id, workflowID, fromNodeID, toNodeID string) *Edge {
	return &Edge{
		BaseEntity: domain.NewBaseEntity(id),
		WorkflowID: workflowID,
		FromNodeID: fromNodeID,
		ToNodeID:   toNodeID,
	}
}

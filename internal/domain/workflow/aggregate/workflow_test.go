package aggregate

import (
	"testing"
	"time"

	"use-open-workflow.io/engine/pkg/id"
)

type mockIDFactory struct {
	counter int
}

func (m *mockIDFactory) New() string {
	m.counter++
	return "mock-id-" + string(rune('0'+m.counter))
}

var _ id.Factory = (*mockIDFactory)(nil)

func TestNewWorkflow_HasTimestamps(t *testing.T) {
	factory := NewWorkflowFactory(&mockIDFactory{})
	before := time.Now().UTC()

	workflow := factory.Make("Test Workflow", "Test Description")

	after := time.Now().UTC()

	if workflow.CreatedAt.Before(before) || workflow.CreatedAt.After(after) {
		t.Errorf("CreatedAt should be between before and after, got %v", workflow.CreatedAt)
	}
	if workflow.UpdatedAt.Before(before) || workflow.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt should be between before and after, got %v", workflow.UpdatedAt)
	}
	if !workflow.CreatedAt.Equal(workflow.UpdatedAt) {
		t.Errorf("CreatedAt and UpdatedAt should be equal on creation")
	}
}

func TestReconstituteWorkflow_PreservesTimestamps(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 6, 15, 18, 30, 0, 0, time.UTC)

	workflow := ReconstituteWorkflow("wf-id", "Test Workflow", "Description", createdAt, updatedAt)

	if !workflow.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt should be %v, got %v", createdAt, workflow.CreatedAt)
	}
	if !workflow.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt should be %v, got %v", updatedAt, workflow.UpdatedAt)
	}
}

func TestUpdateName_UpdatesTimestamp(t *testing.T) {
	factory := &mockIDFactory{}
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	workflow := ReconstituteWorkflow("wf-id", "Original Name", "Description", createdAt, updatedAt)

	before := time.Now().UTC()
	workflow.UpdateName(factory, "New Name")
	after := time.Now().UTC()

	if workflow.UpdatedAt.Before(before) || workflow.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt should be updated to current time, got %v", workflow.UpdatedAt)
	}
}

func TestUpdateName_PreservesCreatedAt(t *testing.T) {
	factory := &mockIDFactory{}
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	workflow := ReconstituteWorkflow("wf-id", "Original Name", "Description", createdAt, updatedAt)

	workflow.UpdateName(factory, "New Name")

	if !workflow.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt should remain %v, got %v", createdAt, workflow.CreatedAt)
	}
}

func TestAddNodeDefinition_AddsToWorkflow(t *testing.T) {
	factory := &mockIDFactory{}
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	config := map[string]interface{}{"key": "value"}
	nodeDef := workflow.AddNodeDefinition(factory, "template-id", "Node 1", config, 100.0, 200.0)

	if len(workflow.NodeDefinitions) != 1 {
		t.Errorf("Expected 1 node definition, got %d", len(workflow.NodeDefinitions))
	}
	if nodeDef.Name != "Node 1" {
		t.Errorf("Expected node name 'Node 1', got %s", nodeDef.Name)
	}
	if nodeDef.NodeTemplateID != "template-id" {
		t.Errorf("Expected template ID 'template-id', got %s", nodeDef.NodeTemplateID)
	}
	if nodeDef.PositionX != 100.0 || nodeDef.PositionY != 200.0 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", nodeDef.PositionX, nodeDef.PositionY)
	}
}

func TestAddEdge_AddsToWorkflow(t *testing.T) {
	factory := &mockIDFactory{}
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	// Add two node definitions first
	nodeDef1 := workflow.AddNodeDefinition(factory, "template-id", "Node 1", nil, 0, 0)
	nodeDef2 := workflow.AddNodeDefinition(factory, "template-id", "Node 2", nil, 100, 0)

	edge, err := workflow.AddEdge(factory, nodeDef1.ID, nodeDef2.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(workflow.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(workflow.Edges))
	}
	if edge.FromNodeDefinitionID != nodeDef1.ID {
		t.Errorf("Expected from node ID %s, got %s", nodeDef1.ID, edge.FromNodeDefinitionID)
	}
	if edge.ToNodeDefinitionID != nodeDef2.ID {
		t.Errorf("Expected to node ID %s, got %s", nodeDef2.ID, edge.ToNodeDefinitionID)
	}
}

func TestAddEdge_FailsForNonExistentFromNode(t *testing.T) {
	factory := &mockIDFactory{}
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	nodeDef := workflow.AddNodeDefinition(factory, "template-id", "Node 1", nil, 0, 0)

	_, err := workflow.AddEdge(factory, "non-existent-id", nodeDef.ID)

	if err == nil {
		t.Error("Expected error for non-existent from node")
	}
}

func TestAddEdge_FailsForNonExistentToNode(t *testing.T) {
	factory := &mockIDFactory{}
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	nodeDef := workflow.AddNodeDefinition(factory, "template-id", "Node 1", nil, 0, 0)

	_, err := workflow.AddEdge(factory, nodeDef.ID, "non-existent-id")

	if err == nil {
		t.Error("Expected error for non-existent to node")
	}
}

func TestRemoveNodeDefinition_RemovesFromWorkflow(t *testing.T) {
	factory := &mockIDFactory{}
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	nodeDef := workflow.AddNodeDefinition(factory, "template-id", "Node 1", nil, 0, 0)

	err := workflow.RemoveNodeDefinition(factory, nodeDef.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(workflow.NodeDefinitions) != 0 {
		t.Errorf("Expected 0 node definitions, got %d", len(workflow.NodeDefinitions))
	}
}

func TestRemoveNodeDefinition_FailsForNonExistent(t *testing.T) {
	factory := &mockIDFactory{}
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	err := workflow.RemoveNodeDefinition(factory, "non-existent-id")

	if err == nil {
		t.Error("Expected error for non-existent node definition")
	}
}

func TestRemoveNodeDefinition_RemovesConnectedEdges(t *testing.T) {
	factory := &mockIDFactory{}
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	nodeDef1 := workflow.AddNodeDefinition(factory, "template-id", "Node 1", nil, 0, 0)
	nodeDef2 := workflow.AddNodeDefinition(factory, "template-id", "Node 2", nil, 100, 0)
	workflow.AddEdge(factory, nodeDef1.ID, nodeDef2.ID)

	if len(workflow.Edges) != 1 {
		t.Errorf("Expected 1 edge before removal, got %d", len(workflow.Edges))
	}

	err := workflow.RemoveNodeDefinition(factory, nodeDef1.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(workflow.Edges) != 0 {
		t.Errorf("Expected 0 edges after removing connected node, got %d", len(workflow.Edges))
	}
}

func TestRemoveEdge_RemovesFromWorkflow(t *testing.T) {
	factory := &mockIDFactory{}
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	nodeDef1 := workflow.AddNodeDefinition(factory, "template-id", "Node 1", nil, 0, 0)
	nodeDef2 := workflow.AddNodeDefinition(factory, "template-id", "Node 2", nil, 100, 0)
	edge, _ := workflow.AddEdge(factory, nodeDef1.ID, nodeDef2.ID)

	err := workflow.RemoveEdge(factory, edge.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(workflow.Edges) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(workflow.Edges))
	}
}

func TestRemoveEdge_FailsForNonExistent(t *testing.T) {
	factory := &mockIDFactory{}
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	err := workflow.RemoveEdge(factory, "non-existent-id")

	if err == nil {
		t.Error("Expected error for non-existent edge")
	}
}

func TestGetNodeDefinition_ReturnsCorrectNode(t *testing.T) {
	factory := &mockIDFactory{}
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	nodeDef := workflow.AddNodeDefinition(factory, "template-id", "Node 1", nil, 0, 0)

	found := workflow.GetNodeDefinition(nodeDef.ID)

	if found == nil {
		t.Error("Expected to find node definition")
	}
	if found.ID != nodeDef.ID {
		t.Errorf("Expected node ID %s, got %s", nodeDef.ID, found.ID)
	}
}

func TestGetNodeDefinition_ReturnsNilForNonExistent(t *testing.T) {
	workflow := ReconstituteWorkflow("wf-id", "Workflow", "Description", time.Now().UTC(), time.Now().UTC())

	found := workflow.GetNodeDefinition("non-existent-id")

	if found != nil {
		t.Error("Expected nil for non-existent node definition")
	}
}

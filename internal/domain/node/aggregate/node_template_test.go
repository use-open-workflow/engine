package aggregate

import (
	"testing"
	"time"

	"use-open-workflow.io/engine/pkg/id"
)

type mockIDFactory struct{}

func (m *mockIDFactory) New() string {
	return "mock-id"
}

var _ id.Factory = (*mockIDFactory)(nil)

func TestNewNodeTemplate_HasTimestamps(t *testing.T) {
	factory := &mockIDFactory{}
	before := time.Now().UTC()

	template := newNodeTemplate(factory, "agg-id", "Test Template")

	after := time.Now().UTC()

	if template.CreatedAt.Before(before) || template.CreatedAt.After(after) {
		t.Errorf("CreatedAt should be between before and after, got %v", template.CreatedAt)
	}
	if template.UpdatedAt.Before(before) || template.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt should be between before and after, got %v", template.UpdatedAt)
	}
	if !template.CreatedAt.Equal(template.UpdatedAt) {
		t.Errorf("CreatedAt and UpdatedAt should be equal on creation")
	}
}

func TestReconstituteNodeTemplate_PreservesTimestamps(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 6, 15, 18, 30, 0, 0, time.UTC)

	template := ReconstituteNodeTemplate("agg-id", "Test Template", createdAt, updatedAt)

	if !template.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt should be %v, got %v", createdAt, template.CreatedAt)
	}
	if !template.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt should be %v, got %v", updatedAt, template.UpdatedAt)
	}
}

func TestUpdateName_UpdatesTimestamp(t *testing.T) {
	factory := &mockIDFactory{}
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	template := ReconstituteNodeTemplate("agg-id", "Original Name", createdAt, updatedAt)

	before := time.Now().UTC()
	template.UpdateName(factory, "New Name")
	after := time.Now().UTC()

	if template.UpdatedAt.Before(before) || template.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt should be updated to current time, got %v", template.UpdatedAt)
	}
}

func TestUpdateName_PreservesCreatedAt(t *testing.T) {
	factory := &mockIDFactory{}
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	template := ReconstituteNodeTemplate("agg-id", "Original Name", createdAt, updatedAt)

	template.UpdateName(factory, "New Name")

	if !template.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt should remain %v, got %v", createdAt, template.CreatedAt)
	}
}

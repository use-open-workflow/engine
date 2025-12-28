package inbound

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNodeTemplateDTO_JSONSerialization(t *testing.T) {
	dto := NodeTemplateDTO{
		ID:        "test-id",
		Name:      "Test Template",
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC),
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("Failed to marshal DTO: %v", err)
	}

	jsonStr := string(data)

	// Check for camelCase field names
	if !contains(jsonStr, `"id"`) {
		t.Error("JSON should contain 'id' field")
	}
	if !contains(jsonStr, `"name"`) {
		t.Error("JSON should contain 'name' field")
	}
	if !contains(jsonStr, `"createdAt"`) {
		t.Error("JSON should contain 'createdAt' field (camelCase)")
	}
	if !contains(jsonStr, `"updatedAt"`) {
		t.Error("JSON should contain 'updatedAt' field (camelCase)")
	}

	// Check RFC 3339 format
	if !contains(jsonStr, "2024-01-15T10:30:00Z") {
		t.Errorf("CreatedAt should be in RFC 3339 format, got: %s", jsonStr)
	}
	if !contains(jsonStr, "2024-06-20T14:45:00Z") {
		t.Errorf("UpdatedAt should be in RFC 3339 format, got: %s", jsonStr)
	}
}

func TestNodeTemplateDTO_JSONDeserialization(t *testing.T) {
	jsonStr := `{"id":"test-id","name":"Test Template","createdAt":"2024-01-15T10:30:00Z","updatedAt":"2024-06-20T14:45:00Z"}`

	var dto NodeTemplateDTO
	err := json.Unmarshal([]byte(jsonStr), &dto)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if dto.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", dto.ID)
	}
	if dto.Name != "Test Template" {
		t.Errorf("Expected Name 'Test Template', got '%s'", dto.Name)
	}

	expectedCreatedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !dto.CreatedAt.Equal(expectedCreatedAt) {
		t.Errorf("Expected CreatedAt %v, got %v", expectedCreatedAt, dto.CreatedAt)
	}

	expectedUpdatedAt := time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC)
	if !dto.UpdatedAt.Equal(expectedUpdatedAt) {
		t.Errorf("Expected UpdatedAt %v, got %v", expectedUpdatedAt, dto.UpdatedAt)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

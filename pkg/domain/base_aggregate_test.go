package domain

import (
	"testing"
	"time"
)

func TestNewBaseAggregate_SetsTimestamps(t *testing.T) {
	before := time.Now().UTC()
	agg := NewBaseAggregate("test-id")
	after := time.Now().UTC()

	if agg.CreatedAt.Before(before) || agg.CreatedAt.After(after) {
		t.Errorf("CreatedAt should be between before and after, got %v", agg.CreatedAt)
	}
	if agg.UpdatedAt.Before(before) || agg.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt should be between before and after, got %v", agg.UpdatedAt)
	}
}

func TestNewBaseAggregate_TimestampsAreEqual(t *testing.T) {
	agg := NewBaseAggregate("test-id")

	if !agg.CreatedAt.Equal(agg.UpdatedAt) {
		t.Errorf("CreatedAt and UpdatedAt should be equal on creation, got CreatedAt=%v, UpdatedAt=%v", agg.CreatedAt, agg.UpdatedAt)
	}
}

func TestNewBaseAggregate_TimestampsAreUTC(t *testing.T) {
	agg := NewBaseAggregate("test-id")

	if agg.CreatedAt.Location() != time.UTC {
		t.Errorf("CreatedAt should be in UTC, got %v", agg.CreatedAt.Location())
	}
	if agg.UpdatedAt.Location() != time.UTC {
		t.Errorf("UpdatedAt should be in UTC, got %v", agg.UpdatedAt.Location())
	}
}

func TestReconstituteBaseAggregate_PreservesTimestamps(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 6, 15, 18, 30, 0, 0, time.UTC)

	agg := ReconstituteBaseAggregate("test-id", createdAt, updatedAt)

	if !agg.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt should be %v, got %v", createdAt, agg.CreatedAt)
	}
	if !agg.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt should be %v, got %v", updatedAt, agg.UpdatedAt)
	}
}

func TestSetUpdatedAt_UpdatesOnlyUpdatedAt(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 6, 15, 18, 30, 0, 0, time.UTC)
	newUpdatedAt := time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC)

	agg := ReconstituteBaseAggregate("test-id", createdAt, updatedAt)
	agg.SetUpdatedAt(newUpdatedAt)

	if !agg.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt should remain %v, got %v", createdAt, agg.CreatedAt)
	}
	if !agg.UpdatedAt.Equal(newUpdatedAt) {
		t.Errorf("UpdatedAt should be %v, got %v", newUpdatedAt, agg.UpdatedAt)
	}
}

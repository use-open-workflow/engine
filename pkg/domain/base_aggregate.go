package domain

import "time"

type BaseAggregate struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	events    []Event
}

func NewBaseAggregate(id string) BaseAggregate {
	now := time.Now().UTC()
	return BaseAggregate{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		events:    make([]Event, 0),
	}
}

func ReconstituteBaseAggregate(id string, createdAt time.Time, updatedAt time.Time) BaseAggregate {
	return BaseAggregate{
		ID:        id,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		events:    make([]Event, 0),
	}
}

func (a *BaseAggregate) SetUpdatedAt(t time.Time) {
	a.UpdatedAt = t
}

func (a *BaseAggregate) AddEvent(event Event) {
	a.events = append(a.events, event)
}

func (a *BaseAggregate) Events() []Event {
	return a.events
}

func (a *BaseAggregate) ClearEvents() {
	a.events = make([]Event, 0)
}

func (*BaseAggregate) IsAggregate() bool {
	return true
}

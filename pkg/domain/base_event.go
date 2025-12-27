package domain

import "time"

type BaseEvent struct {
	id            string
	aggregateID   string
	aggregateType string
	eventType     string
	occurredAt    time.Time
	version       int
}

func NewBaseEvent(id, aggregateID, aggregateType, eventType string) BaseEvent {
	return BaseEvent{
		id:            id,
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		eventType:     eventType,
		occurredAt:    time.Now().UTC(),
		version:       1,
	}
}

func (e BaseEvent) ID() string            { return e.id }
func (e BaseEvent) AggregateID() string   { return e.aggregateID }
func (e BaseEvent) AggregateType() string { return e.aggregateType }
func (e BaseEvent) EventType() string     { return e.eventType }
func (e BaseEvent) OccurredAt() time.Time { return e.occurredAt }
func (e BaseEvent) Version() int          { return e.version }
func (BaseEvent) IsEvent() bool           { return true }

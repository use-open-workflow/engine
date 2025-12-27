package domain

import "time"

type Aggregate interface {
	AddEvent(event Event)
	Events() []Event
	ClearEvents()
	IsAggregate() bool
}

type Entity interface {
	IsEntity() bool
}

type Event interface {
	IsEvent() bool
	ID() string
	AggregateID() string
	AggregateType() string
	EventType() string
	OccurredAt() time.Time
	Version() int
}

type ValueObject interface {
	IsValueObject() bool
}

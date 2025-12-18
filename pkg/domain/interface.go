package domain

type Aggregate interface {
	IsAggregate() bool
}

type Entity interface {
	IsEntity() bool
}

type Event interface {
	IsEvent() bool
}

type ValueObject interface {
	IsValueObject() bool
}

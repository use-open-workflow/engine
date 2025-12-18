package domain

type BaseAggregate struct {
	ID     string
	events []Event
}

func NewBaseAggregate(id string) BaseAggregate {
	return BaseAggregate{
		ID:     id,
		events: []Event{},
	}
}

func (a BaseAggregate) AddEvent(event Event) {
	a.events = append(a.events, event)
}

func (BaseAggregate) IsAggregate() bool {
	return true
}

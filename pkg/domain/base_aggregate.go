package domain

type BaseAggregate struct {
	ID     string
	events []Event
}

func NewBaseAggregate(id string) BaseAggregate {
	return BaseAggregate{
		ID:     id,
		events: make([]Event, 0),
	}
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

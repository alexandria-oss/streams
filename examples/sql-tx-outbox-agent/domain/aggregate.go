package domain

type Aggregate interface {
	PullEvents() []Event
	PushEvents(e ...Event)
}

type BaseAggregate struct {
	events []Event
}

var _ Aggregate = &BaseAggregate{}

func (b *BaseAggregate) PullEvents() []Event {
	return b.events
}

func (b *BaseAggregate) PushEvents(e ...Event) {
	b.events = append(b.events, e...)
}

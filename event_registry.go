package streams

import (
	"github.com/modern-go/reflect2"
)

// A EventRegistry is a low-level storage used to create relationships between Event types and topics (streams).
type EventRegistry map[string]string

// RegisterEvent creates a relationship between an Event type and a topic (stream).
func (r EventRegistry) RegisterEvent(event Event, topic string) {
	typeOf := reflect2.TypeOf(event)
	r[typeOf.String()] = topic
}

// GetEventTopic retrieves the attached topic of the Event. Returns ErrEventNotFound if Event entry is not available.
func (r EventRegistry) GetEventTopic(event Event) (string, error) {
	// NOTE: Might require to use a hash-bag instead to keep lookups for both event type and topics in constant time range.
	typeOf := reflect2.TypeOf(event)
	out, ok := r[typeOf.String()]
	if !ok {
		return out, ErrEventNotFound
	}
	return out, nil
}

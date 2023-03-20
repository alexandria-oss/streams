package streams

// An Event is an abstraction unit of Message. Represents factual information about a happening inside a system
// (hence its immutable).
type Event interface {
	// GetHeaders allocates a set of key-value items to be passed through data-in-motion platforms as message headers.
	GetHeaders() map[string]string
	// GetKey retrieves a key used by certain data-in-motion platforms (e.g. Apache Kafka) to route
	// messages with the same key to certain partitions/queues. Leave empty if routing is NOT desired.
	GetKey() string
}

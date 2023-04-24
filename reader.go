package streams

import "context"

// A ReadTask is the unit of information a SubscriberScheduler passes to Reader workers in order to start
// stream-reading jobs. Use ExternalArgs to specify driver-specific configuration.
type ReadTask struct {
	Stream       string
	Handler      ReaderHandleFunc
	ExternalArgs map[string]any
}

// SetArg sets an entry into ExternalArgs and returns the ReadTask instance ready to be chained to another builder
// routine (Fluent API-like). Arguments will be later passed to Reader concrete implementations; useful for situations
// where readers accept extra arguments such as consumer groups identifiers (Apache Kafka).
func (t *ReadTask) SetArg(key string, value any) *ReadTask {
	if t.ExternalArgs == nil {
		t.ExternalArgs = make(map[string]any)
	}
	t.ExternalArgs[key] = value
	return t
}

// WithMiddleware appends a ReaderHandleFunc instance to ReadTask.Handler; this is also known as
// chain of responsibility pattern.
func (t *ReadTask) WithMiddleware(middlewareFunc ReaderMiddlewareFunc) *ReadTask {
	t.Handler = middlewareFunc(t.Handler)
	return t
}

// A Reader is a low-level component which allows systems to read from a stream.
type Reader interface {
	// Read reads from the specified stream in ReadTask, blocking the I/O. Everytime a new message arrives,
	// Reader will execute ReadTask.Handler routine in a separate goroutine.
	//
	// Use ctx context.Context to signal shutdowns.
	Read(ctx context.Context, task ReadTask) error
}

// ReaderHandleFunc routine to be executed for each message received by Reader instances.
type ReaderHandleFunc func(ctx context.Context, msg Message) error

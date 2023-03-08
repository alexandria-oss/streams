package streams

import "context"

// Reader stream-listening task scheduler.
type Reader interface {
	Read(ctx context.Context, stream string) error
}

// ReaderHandleFunc function to be executed for each message received by Reader instances.
type ReaderHandleFunc func(ctx context.Context, msg Message) error

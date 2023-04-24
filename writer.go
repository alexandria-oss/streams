package streams

import (
	"context"
)

// A Writer is a low-level component that writes messages into streams. Each message MUST have its own stream name specified.
type Writer interface {
	// Write writes a message batch into a stream specified on each Message through the Message.StreamName field.
	//
	// Depending on the underlying Writer implementation, this routine will write messages in actual batches,
	// batch chunks or one-by-one.
	Write(ctx context.Context, msgBatch []Message) error
}

// NoopWriter no-operation Writer instance.
type NoopWriter struct {
	WantWriterErr error
}

var _ Writer = NoopWriter{}

func (n NoopWriter) Write(_ context.Context, _ []Message) error {
	return n.WantWriterErr
}

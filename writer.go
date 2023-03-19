package streams

import (
	"context"
)

// A Writer is a low-level component that writes into streams.
type Writer interface {
	Write(ctx context.Context, msgBatch []Message) error
}

type NoopWriter struct {
	WantWriterErr error
}

var _ Writer = NoopWriter{}

func (n NoopWriter) Write(_ context.Context, _ []Message) error {
	return n.WantWriterErr
}

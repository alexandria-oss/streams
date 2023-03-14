package chanbuf

import (
	"context"

	"github.com/alexandria-oss/streams"
)

type Writer struct {
	bus *Bus
}

var _ streams.Writer = Writer{}

func NewWriter(b *Bus) Writer {
	if b == nil {
		newDefaultBus()
		b = DefaultBus
	}

	return Writer{
		bus: b,
	}
}

func (w Writer) Write(_ context.Context, msgBatch []streams.Message) error {
	for _, msg := range msgBatch {
		if err := w.bus.Publish(msg); err != nil {
			return err
		}
	}
	return nil
}

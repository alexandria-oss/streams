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
		b = defaultBus
	}

	return Writer{
		bus: b,
	}
}

func (w Writer) Write(ctx context.Context, msgBatch []streams.Message) error {
	//TODO implement me
	panic("implement me")
}

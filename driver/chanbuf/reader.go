package chanbuf

import (
	"context"

	"github.com/alexandria-oss/streams"
)

type Reader struct {
	bus *Bus
}

var _ streams.Reader = Reader{}

func NewReader(b *Bus) Reader {
	if b == nil {
		newDefaultBus()
		b = DefaultBus
	}

	return Reader{
		bus: b,
	}
}

func (r Reader) Read(_ context.Context, task streams.ReadTask) error {
	r.bus.Subscribe(task.Stream, task.Handler)
	return nil
}

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

func (r Reader) Read(_ context.Context, stream string, handler streams.ReaderHandleFunc) error {
	r.bus.Subscribe(stream, handler)
	return nil
}

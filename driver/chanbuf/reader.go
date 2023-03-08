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
		b = defaultBus
	}

	return Reader{
		bus: b,
	}
}

func (r Reader) Read(ctx context.Context, stream string) error {
	//TODO implement me
	panic("implement me")
}

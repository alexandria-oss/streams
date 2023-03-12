package streams

import (
	"context"
)

type Writer interface {
	Write(ctx context.Context, msgBatch []Message) error
}

type WriterConfig struct {
	IdentifierFactory IdentifierFactory
}

var DefaultWriterConfig = WriterConfig{
	IdentifierFactory: NewKSUID,
}

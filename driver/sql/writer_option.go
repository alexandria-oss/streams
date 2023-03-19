package sql

import (
	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/codec"
)

// A WriterOption is used to configure a Writer instance in an idiomatic & fine-grained way.
type WriterOption interface {
	apply(*WriterConfig)
}

type egressTableOption struct {
	table string
}

var _ WriterOption = egressTableOption{}

func (e egressTableOption) apply(config *WriterConfig) {
	config.WriterEgressTable = e.table
}

// WithEgressTable sets the name of the table to be used as <<message egress table>>. A <<message egress table>> is
// a system database table used by `streams` mechanisms to write batch of messages to be published into a message stream.
func WithEgressTable(table string) WriterOption {
	return egressTableOption{table: table}
}

type codecOption struct {
	codec codec.Codec
}

var _ WriterOption = codecOption{}

func (o codecOption) apply(opts *WriterConfig) {
	opts.Codec = o.codec
}

// WithCodec sets the codec.Codec to be used by Writer to encode message batches, so data may be stored into
// a database efficiently.
func WithCodec(c codec.Codec) WriterOption {
	return codecOption{codec: c}
}

type identifierFactoryOption struct {
	identifierFactory streams.IdentifierFactory
}

var _ WriterOption = identifierFactoryOption{}

func (o identifierFactoryOption) apply(opts *WriterConfig) {
	opts.IdentifierFactory = o.identifierFactory
}

// WithIdentifierFactory sets the IdentifierFactory algorithm used by inner processes when a unique identifier
// is required (e.g. generate an id for an incoming message batch).
func WithIdentifierFactory(f streams.IdentifierFactory) WriterOption {
	return identifierFactoryOption{identifierFactory: f}
}

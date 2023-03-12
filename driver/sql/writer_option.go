package sql

import "github.com/alexandria-oss/streams/codec"

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

func WithCodec(c codec.Codec) WriterOption {
	return codecOption{codec: c}
}

package streams

import "github.com/alexandria-oss/streams/codec"

type PublisherOption interface {
	apply(config *PublisherConfig)
}

type publisherIdFactory struct {
	factory IdentifierFactory
}

var _ PublisherOption = publisherIdFactory{}

func (p publisherIdFactory) apply(config *PublisherConfig) {
	config.IdentifierFactory = p.factory
}

func WithIdentifierFactory(f IdentifierFactory) PublisherOption {
	return publisherIdFactory{factory: f}
}

type publisherCodec struct {
	c codec.Codec
}

var _ PublisherOption = publisherCodec{}

func (p publisherCodec) apply(config *PublisherConfig) {
	config.Codec = p.c
}

func WithPublisherCodec(c codec.Codec) PublisherOption {
	return publisherCodec{c: c}
}

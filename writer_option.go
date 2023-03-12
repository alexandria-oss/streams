package streams

type WriterOption interface {
	ApplyWriterCfg(*WriterConfig)
}

type NoopWriterOption struct{}

var _ WriterOption = NoopWriterOption{}

func (o NoopWriterOption) ApplyWriterCfg(_ *WriterConfig) {}

type identifierFactoryOption struct {
	identifierFactory IdentifierFactory
}

var _ WriterOption = identifierFactoryOption{}

func (o identifierFactoryOption) ApplyWriterCfg(opts *WriterConfig) {
	opts.IdentifierFactory = o.identifierFactory
}

func WithIdentifierFactory(f IdentifierFactory) WriterOption {
	return identifierFactoryOption{identifierFactory: f}
}

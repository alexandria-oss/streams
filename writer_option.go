package streams

// A WriterOption is used to configure a Writer instance in an idiomatic & fine-grained way.
type WriterOption interface {
	// ApplyWriterCfg applies the configuration from the current WriterOperation state into WriterConfig pointer
	// reference.
	ApplyWriterCfg(*WriterConfig)
}

// A NoopWriterOption is a no-operation Writer instance.
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

// WithIdentifierFactory sets the IdentifierFactory algorithm used by inner processes when a unique identifier
// is required.
func WithIdentifierFactory(f IdentifierFactory) WriterOption {
	return identifierFactoryOption{identifierFactory: f}
}

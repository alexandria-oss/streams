package streams

// A Bus is the top-level component used by systems to interact with data-in-motion platforms.
// If finer-grained tuning or experience is desired, use low-level components such as Writer or Reader.
type Bus struct {
	Publisher
	SubscriberScheduler
	EventRegistry
}

// NewBus allocates a Bus instance. Specify options to customize Publisher's default configurations.
func NewBus(w Writer, r Reader, options ...PublisherOption) Bus {
	reg := make(EventRegistry)
	return Bus{
		EventRegistry:       reg,
		Publisher:           NewPublisher(w, reg, options...),
		SubscriberScheduler: NewSubscriberScheduler(r, reg),
	}
}

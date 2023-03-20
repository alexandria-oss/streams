package egress

// A Notifier is an egress proxy component used by a system to notify the egress proxy agent
// to forward traffic buffer (batch) into a stream.
type Notifier interface {
	// NotifyAgent triggers egress proxy agent to forward traffic buffer (batch) into a stream.
	NotifyAgent(batchID string) error
}

// EmbeddedNotifier an agent-less Notifier implementation used by a system to call a Forwarder instance
// directly.
type EmbeddedNotifier struct {
	Forwarder Forwarder
}

var _ Notifier = EmbeddedNotifier{}

func (n EmbeddedNotifier) NotifyAgent(batchID string) error {
	return n.Forwarder.Forward(batchID)
}

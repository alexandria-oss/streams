package listener

import (
	"context"
	"errors"

	"github.com/alexandria-oss/streams/proxy/egress"
)

var ErrUnknownDriver = errors.New("streams_egress_proxy_agent: unknown listener driver")

type Listener interface {
	Start() error
	Close(ctx context.Context) error
}

func NewListener(driver string, fwd egress.Forwarder) (Listener, error) {
	switch driver {
	case WALDriver:
		return NewWAL(fwd), nil
	default:
		return nil, ErrUnknownDriver
	}
}

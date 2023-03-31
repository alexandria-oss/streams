package listener

import (
	"context"
	"errors"
)

var ErrUnknownDriver = errors.New("streams_egress_proxy_agent: unknown listener driver")

type Listener interface {
	Start() error
	Close(ctx context.Context) error
}

func NewListener(driver string) (Listener, error) {
	switch driver {
	case WALDriver:
		return NewWAL(), nil
	default:
		return nil, ErrUnknownDriver
	}
}

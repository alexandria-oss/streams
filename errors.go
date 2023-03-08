package streams

import "errors"

var (
	ErrBusIsShutdown = errors.New("streams: bus has been terminated")
	ErrEmptyMessage  = errors.New("streams: message is empty")
)

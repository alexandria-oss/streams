package streams

import (
	"errors"
	"fmt"
)

var (
	ErrBusIsShutdown = errors.New("streams: bus has been terminated")
	ErrEmptyMessage  = errors.New("streams: message is empty")
	ErrUnrecoverable = errors.New("streams: unrecoverable error")
)

// A ErrUnrecoverableWrap is a special wrapper for certain type of errors with no recoverable action.
type ErrUnrecoverableWrap struct {
	ParentErr error
}

var _ error = ErrUnrecoverableWrap{}

var _ fmt.Stringer = ErrUnrecoverableWrap{}

func (u ErrUnrecoverableWrap) String() string {
	return u.Error()
}

func (u ErrUnrecoverableWrap) Error() string {
	return u.ParentErr.Error()
}

// Unwrap returns ErrUnrecoverable variable. Thus, calls to errors.Is(err, ErrUnrecoverable) routine will detect this
// wrapper as unrecoverable error.
func (u ErrUnrecoverableWrap) Unwrap() error {
	return ErrUnrecoverable
}

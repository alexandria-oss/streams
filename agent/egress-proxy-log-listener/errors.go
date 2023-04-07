package egress_proxy_log_listener

import "fmt"

type FatalError struct {
	Err error
}

var _ error = FatalError{}

var _ fmt.Stringer = FatalError{}

func (f FatalError) Error() string {
	return f.Err.Error()
}

func (f FatalError) String() string {
	return f.Error()
}

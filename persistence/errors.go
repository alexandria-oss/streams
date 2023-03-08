package persistence

import "errors"

var (
	ErrTransactionContextNotFound = errors.New("streams: transaction context not found")
)

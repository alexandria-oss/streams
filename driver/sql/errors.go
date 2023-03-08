package sql

import "errors"

var (
	ErrUnableToWriteRows   = errors.New("streams.sql: unable to write rows")
	ErrTransactionNotFound = errors.New("streams.sql: transaction not found")
)

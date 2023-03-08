package sql

import "context"

type TransactionContextKeyType string

const TransactionContextKey TransactionContextKeyType = "streams.tx_context"

func SetTransactionContext[T any](ctx context.Context, tx T) context.Context {
	return context.WithValue(ctx, TransactionContextKey, tx)
}

func GetTransactionContext[T any](ctx context.Context) (T, error) {
	tx, ok := ctx.Value(TransactionContextKey).(T)
	if !ok {
		var zeroVal T
		return zeroVal, ErrTransactionNotFound
	}
	return tx, nil
}

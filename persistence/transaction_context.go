package persistence

import "context"

// TransactionContextKeyType custom type used by transaction contexts.
type TransactionContextKeyType string

// TransactionContextKey context key used by a transaction context.
const TransactionContextKey TransactionContextKeyType = "streams.tx_context"

// SetTransactionContext allocates a transaction context using ctx as parent.
//
// A transaction context is a mechanism used by multiple system layers to share one single transaction.
// Thus, each operation inside the same context will be part of the transaction itself.
//
// Example:
//
// txCtx := SetTransactionContext[*sql.Tx](context.TODO(), mySqlTx)
//
// repository.DoOp(txCtx) // this will be part of the transaction as well
//
// in repository.DoOp(ctx context.Context)
//
// tx, err := GetTransactionContext[*sql.Tx](ctx)
//
// tx.Exec()...
func SetTransactionContext[T any](ctx context.Context, tx T) context.Context {
	return context.WithValue(ctx, TransactionContextKey, tx)
}

// GetTransactionContext retrieves a transaction context from ctx.
//
// A transaction context is a mechanism used by multiple system layers to share one single transaction.
// Thus, each operation inside the same context will be part of the transaction itself.
//
// Example:
//
// txCtx := SetTransactionContext[*sql.Tx](context.TODO(), mySqlTx)
//
// repository.DoOp(txCtx) // this will be part of the transaction as well
//
// in repository.DoOp(ctx context.Context)
//
// tx, err := GetTransactionContext[*sql.Tx](ctx)
//
// tx.Exec()...
func GetTransactionContext[T any](ctx context.Context) (T, error) {
	tx, ok := ctx.Value(TransactionContextKey).(T)
	if !ok {
		var zeroVal T
		return zeroVal, ErrTransactionContextNotFound
	}
	return tx, nil
}

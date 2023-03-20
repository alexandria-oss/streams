package persistence

import "context"

// TransactionContextKeyType custom type used by transaction contexts.
type TransactionContextKeyType string

// TransactionContextKey context key used by a transaction context.
const TransactionContextKey TransactionContextKeyType = "streams.tx_context"

// A TransactionContext is a mechanism used by multiple system layers to share one single transaction.
// Thus, each operation inside the same context will be part of the transaction itself.
type TransactionContext[T any] struct {
	TransactionID string
	Tx            T
}

// SetTransactionContext allocates a transaction context using ctx as parent.
//
// A transaction context is a mechanism used by multiple system layers to share one single transaction.
// Thus, each operation inside the same context will be part of the transaction itself.
//
// Example:
//
// txCtx := SetTransactionContext(context.TODO(), TransactionContext[*sql.Tx]{TransactionID: "123", Tx: mySqlTx})
//
// repository.DoOp(txCtx) // this will be part of the transaction as well
//
// in repository.DoOp(ctx context.Context)
//
// tx, err := GetTransactionContext[*sql.Tx](ctx)
//
// tx.Exec()...
func SetTransactionContext[T any](ctx context.Context, tx TransactionContext[T]) context.Context {
	return context.WithValue(ctx, TransactionContextKey, tx)
}

// GetTransactionContext retrieves a transaction context from ctx.
//
// A transaction context is a mechanism used by multiple system layers to share one single transaction.
// Thus, each operation inside the same context will be part of the transaction itself.
//
// Example:
//
// txCtx := SetTransactionContext(context.TODO(), TransactionContext[*sql.Tx]{TransactionID: "123", Tx: mySqlTx})
//
// repository.DoOp(txCtx) // this will be part of the transaction as well
//
// in repository.DoOp(ctx context.Context)
//
// txCtx, err := GetTransactionContext[*sql.Tx](ctx)
//
// txCtx.Tx.Exec()...
func GetTransactionContext[T any](ctx context.Context) (TransactionContext[T], error) {
	tx, ok := ctx.Value(TransactionContextKey).(TransactionContext[T])
	if !ok {
		var zeroVal TransactionContext[T]
		return zeroVal, ErrTransactionContextNotFound
	}
	return tx, nil
}

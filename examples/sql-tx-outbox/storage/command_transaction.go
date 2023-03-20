package storage

import (
	"context"
	"database/sql"
	"sample/domain"

	"github.com/alexandria-oss/streams/proxy/egress"

	"github.com/alexandria-oss/streams"

	"github.com/alexandria-oss/streams/persistence"
)

func WithEmbeddedEgressAgent(args any, next domain.CommandHandlerFunc) domain.CommandHandlerFunc {
	notifier := args.(egress.Notifier)
	return func(ctx context.Context, cmd any) error {
		txID, _ := streams.NewKSUID()
		scopedCtx := persistence.SetTransactionContext(ctx, persistence.TransactionContext[any]{
			TransactionID: txID,
		})
		if err := next(scopedCtx, cmd); err != nil {
			return err
		}

		return notifier.NotifyAgent(txID)
	}
}

func WithSQLTransaction(args any, next domain.CommandHandlerFunc) domain.CommandHandlerFunc {
	db := args.(*sql.DB)
	return func(ctx context.Context, cmd any) error {
		conn, err := db.Conn(ctx)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := conn.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
		}()

		tx, err := conn.BeginTx(ctx, &sql.TxOptions{
			Isolation: 0,
			ReadOnly:  false,
		})

		parentCtx, err := persistence.GetTransactionContext[any](ctx)
		if err != nil {
			return err
		}

		scopedCtx := persistence.SetTransactionContext(ctx, persistence.TransactionContext[*sql.Tx]{
			TransactionID: parentCtx.TransactionID,
			Tx:            tx,
		})
		if nextErr := next(scopedCtx, cmd); nextErr != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return nextErr
		}

		return tx.Commit()
	}
}

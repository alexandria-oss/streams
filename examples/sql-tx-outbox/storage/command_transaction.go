package storage

import (
	"context"
	"database/sql"
	"sample/domain"

	"github.com/alexandria-oss/streams/persistence"
)

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

		scopedCtx := persistence.SetTransactionContext[*sql.Tx](ctx, tx)
		if nextErr := next(scopedCtx, cmd); nextErr != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return rollbackErr
			}
			return nextErr
		}

		return tx.Commit()
	}
}

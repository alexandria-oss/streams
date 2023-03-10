package payment

import (
	"context"
	"database/sql"

	"github.com/alexandria-oss/streams/persistence"
)

// TODO: Move this middleware to command bus when exec handler (cmds using decorators?). That will centralize transaction context usage.

type ServiceTransactionContext struct {
	db   *sql.DB
	next Service
}

var _ Service = ServiceTransactionContext{}

func NewServiceTransactionContext(db *sql.DB, next Service) ServiceTransactionContext {
	return ServiceTransactionContext{
		db:   db,
		next: next,
	}
}

func (s ServiceTransactionContext) Create(ctx context.Context, userID string, amount float64) (Payment, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return Payment{}, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})

	scopedCtx := persistence.SetTransactionContext[*sql.Tx](ctx, tx)
	payment, nextErr := s.next.Create(scopedCtx, userID, amount)
	if nextErr != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return Payment{}, rollbackErr
		}
		return Payment{}, nextErr
	}

	return payment, tx.Commit()
}

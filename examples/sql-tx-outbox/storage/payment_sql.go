package storage

import (
	"context"
	"database/sql"
	"sample/domain/payment"

	"github.com/alexandria-oss/streams/persistence"
)

type PaymentSQL struct {
}

var _ payment.Repository = PaymentSQL{}

func (p PaymentSQL) Save(ctx context.Context, data payment.Payment) error {
	tx, err := persistence.GetTransactionContext[*sql.Tx](ctx)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO payments(payment_id, user_id, amount) VALUES ($1,$2,$3)",
		data.ID, data.UserID, data.Amount)
	if err != nil {
		return err
	}

	//_, err = tx.ExecContext(ctx, "INSERT INTO user_payments_stats(user_id, amount) VALUES ($1,$2)",
	//	data.UserID, data.Amount*2) // only for tx testing purposes
	return err
}

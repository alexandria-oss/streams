package payment

import "sample/domain"

const (
	StreamName = "org.alexandria.payments"
)

type Payment struct {
	*domain.BaseAggregate

	ID     string  `json:"payment_id"`
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

var _ domain.Aggregate = &Payment{}

package payment

import "sample/domain"

type Event struct {
	PaymentID     string  `json:"payment_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	OperationType string
}

var _ domain.Event = Event{}

func (e Event) GetKey() string {
	return e.PaymentID
}

func (e Event) GetStreamName() string {
	return StreamName
}

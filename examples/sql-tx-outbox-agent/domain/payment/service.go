package payment

import (
	"context"
	"sample/domain"
	"sample/messaging"

	"github.com/alexandria-oss/streams"
	"github.com/google/uuid"
)

type Service struct {
	repo        Repository
	eventWriter streams.Writer // TODO: Use a higher-level component instead writer such as bus
}

var DefaultService Service

func NewService(repo Repository, w streams.Writer) Service {
	return Service{
		repo:        repo,
		eventWriter: w,
	}
}

func (s Service) Create(ctx context.Context, userID string, amount float64) error {
	payment := Payment{
		ID:            uuid.NewString(),
		UserID:        userID,
		Amount:        amount,
		BaseAggregate: &domain.BaseAggregate{},
	}

	payment.PushEvents(Event{
		PaymentID:     payment.ID,
		UserID:        payment.UserID,
		Amount:        payment.Amount,
		OperationType: "CREATE",
	})

	if err := s.repo.Save(ctx, payment); err != nil {
		return err
	}

	return messaging.WriteEvents(ctx, s.eventWriter, payment)
}

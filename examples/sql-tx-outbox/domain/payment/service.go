package payment

import (
	"context"
	"sample/domain"
	"sample/messaging"

	"github.com/alexandria-oss/streams"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, userID string, amount float64) (Payment, error)
}

type ServiceImpl struct {
	repo        Repository
	eventWriter streams.Writer // TODO: Use a higher-level component instead writer such as bus
}

var _ Service = ServiceImpl{}

func NewService(repo Repository, w streams.Writer) ServiceImpl {
	return ServiceImpl{
		repo:        repo,
		eventWriter: w,
	}
}

func (s ServiceImpl) Create(ctx context.Context, userID string, amount float64) (Payment, error) {
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
		return Payment{}, err
	}

	return payment, messaging.WriteEvents(ctx, s.eventWriter, payment)
}

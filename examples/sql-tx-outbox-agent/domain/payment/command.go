package payment

import (
	"context"
	"sample/domain"
)

type CreateCommand struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

var _ domain.CommandHandlerFunc = HandleCreateCommand

func HandleCreateCommand(ctx context.Context, args any) error {
	command := args.(CreateCommand)
	return DefaultService.Create(ctx, command.UserID, command.Amount)
}

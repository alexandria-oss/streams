package domain

import (
	"context"
)

type Repository[T Aggregate] interface {
	Save(ctx context.Context, data T) error
}

package ingress

import (
	"context"
)

type Storage interface {
	Commit(ctx context.Context, messageID string) error
}

package streams

import "context"

type Writer interface {
	Write(ctx context.Context, msgBatch []Message) error
}

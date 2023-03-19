package streams

import (
	"context"
	"log"

	"github.com/eapache/go-resiliency/retrier"
)

type ReaderMiddlewareFunc func(next ReaderHandleFunc) ReaderHandleFunc

func WithReaderRetry(retry *retrier.Retrier) ReaderMiddlewareFunc {
	return func(next ReaderHandleFunc) ReaderHandleFunc {
		return func(ctx context.Context, msg Message) error {
			return retry.RunCtx(ctx, func(ctx context.Context) error {
				return next(ctx, msg)
			})
		}
	}
}

func WithReaderErrorLogger(logger *log.Logger) ReaderMiddlewareFunc {
	return func(next ReaderHandleFunc) ReaderHandleFunc {
		return func(ctx context.Context, msg Message) error {
			if err := next(ctx, msg); err != nil {
				logger.Print(err)
				return err
			}
			return nil
		}
	}
}

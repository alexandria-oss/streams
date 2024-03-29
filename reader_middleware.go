package streams

import (
	"context"
	"log"

	"github.com/eapache/go-resiliency/retrier"
)

type ReaderMiddlewareFunc func(next ReaderHandleFunc) ReaderHandleFunc

// WithReaderRetry appends to ReaderHandleFunc(s) a mechanism to retry up to N times. Uses retrier.Retrier
// package to enable advanced backoff mechanisms such as exponential plus jitter.
func WithReaderRetry(retry *retrier.Retrier) ReaderMiddlewareFunc {
	return func(next ReaderHandleFunc) ReaderHandleFunc {
		return func(ctx context.Context, msg Message) error {
			return retry.RunCtx(ctx, func(ctx context.Context) error {
				return next(ctx, msg)
			})
		}
	}
}

// WithReaderErrorLogger appends to ReaderHandleFunc(s) a mechanism to log errors using a logger instance.
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

// WithDeduplication appends to ReaderHandleFunc(s) a mechanism to deduplicate processed messages, ensuring
// idempotency. Uses DeduplicationStorage to keep track of processed messages and group identifier
// to enable multiple isolated workers commit their processed messages to the same storage.
func WithDeduplication(group string, storage DeduplicationStorage) ReaderMiddlewareFunc {
	return func(next ReaderHandleFunc) ReaderHandleFunc {
		return func(ctx context.Context, msg Message) error {
			isDupe, err := storage.IsDuplicated(ctx, msg.Headers[group], msg.ID)
			if err != nil {
				return err
			} else if isDupe {
				return nil // ensure idempotency
			}

			if err = next(ctx, msg); err != nil {
				return err
			}

			// no need to fail due a commit message. If the error is sent,
			// this routine could get retried and thus, we would duplicate processes, breaking
			// deduplication guarantees.
			// Concrete implementation should log its own errors anyway.
			storage.Commit(ctx, group, msg.ID)
			return nil
		}
	}
}

// WithDeadLetterQueue appends to ReaderHandleFunc(s) a mechanism to send poisoned or failed (after retries) messages
// to a dead-letter queue (DLQ). The dead-letter queue MIGHT retain these messages for a longer time that a
// normal queue.
//
// Dead-letter queue messages are emitted to the Message.StreamName but with the suffix ".dlq".
//
// After failures, a dead-letter queue comes into play as engineering teams can manually/automatically
// enqueue failed messages again into the original queue (i.e. re-drive/replay policies), so messages can be processed
// again without further overhead.
//
// Moreover, this dead-letter queue could not only be a message bus like Apache Kafka or services like Amazon SQS;
// even a blob storage service like Amazon S3 could implement Writer and retain failed messages.
func WithDeadLetterQueue(writer Writer) ReaderMiddlewareFunc {
	return func(next ReaderHandleFunc) ReaderHandleFunc {
		return func(ctx context.Context, msg Message) error {
			if err := next(ctx, msg); err == nil {
				return nil
			}

			msg.StreamName += ".dlq"
			return writer.Write(ctx, []Message{msg})
		}
	}
}

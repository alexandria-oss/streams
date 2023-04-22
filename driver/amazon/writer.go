package amazon

import (
	"context"
	"sync"

	"github.com/alexandria-oss/streams"
	"github.com/hashicorp/go-multierror"
)

// WriteFunc Amazon service-agnostic message writing function. A Writer instance will call this function
// which is implemented by actual drivers (Amazon SNS/SQS).
type WriteFunc func(ctx context.Context, stream string, msgBatch []streams.Message) error

// A Writer is a generic message writer for Amazon services.
// This component centralizes a batching group buffering algorithm which
// later schedules and executes message writing tasks concurrently to increase write throughput.
//
// This type is NOT ready for usage as standalone component, concrete writers (e.g. sns.Writer, sqs.Writer) should
// be used instead.
type Writer struct {
	WriteFunc WriteFunc
}

var _ streams.Writer = Writer{}

func (w Writer) Write(ctx context.Context, msgBatch []streams.Message) error {
	batchBuf := make(map[string][]streams.Message, len(msgBatch))
	for _, msg := range msgBatch {
		buf, ok := batchBuf[msg.StreamName]
		if !ok {
			buf = make([]streams.Message, 0, 1)
		}

		batchBuf[msg.StreamName] = append(buf, msg)
	}

	errs := &multierror.Error{}
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(batchBuf))
	for stream, msgBuf := range batchBuf {
		go func(streamCopy string, batchCopy []streams.Message) {
			defer wg.Done()
			if err := w.WriteFunc(ctx, streamCopy, batchCopy); err != nil {
				mu.Lock() // multi error is not concurrent safe
				errs = multierror.Append(err, errs)
				mu.Unlock()
			}
		}(stream, msgBuf)
	}
	wg.Wait()
	return errs.ErrorOrNil()
}

package sqs

import (
	"context"
	"errors"
	"strings"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/amazon"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/hashicorp/go-multierror"
)

// WriterConfig is the configuration schema for Amazon SQS streams.Writer implementation.
type WriterConfig struct {
	amazon.Config
	// The length of time, in seconds, for which a specific message is delayed. Valid
	// values: 0 to 900. Maximum: 15 minutes. Messages with a positive DelaySeconds
	// value become available for processing after the delay period is finished. If you
	// don't specify a value, the default value for the queue is applied. When you set
	// FifoQueue , you can't set DelaySeconds per message. You can set this parameter
	// only on a queue level.
	DelaySeconds int32
}

// Writer is the Amazon Simple Queue Service (SQS) streams.Writer implementation.
type Writer struct {
	amazon.Writer
	config WriterConfig

	baseQueueURL string
	client       *sqs.Client
}

var _ streams.Writer = Writer{}

// NewWriter allocates an Amazon Simple Queue Service (SQS) concrete implementation of streams.Writer.
func NewWriter(cfg WriterConfig, awsCfg aws.Config, client *sqs.Client) Writer {
	w := Writer{
		config:       cfg,
		baseQueueURL: newBaseQueueURL(cfg.Config, awsCfg),
		client:       client,
	}
	w.WriteFunc = w.write
	return w
}

func (w Writer) write(ctx context.Context, stream string, msgBatch []streams.Message) error {
	isQueueFIFO := strings.HasSuffix(stream, ".fifo")
	queueURL := newQueueURL(w.baseQueueURL, stream)
	batchBuf := make([]types.SendMessageBatchRequestEntry, len(msgBatch))
	for i, msg := range msgBatch {
		msgID := aws.String(msg.ID)
		entry := types.SendMessageBatchRequestEntry{
			Id:                      msgID,
			MessageBody:             aws.String(string(msg.Data)),
			DelaySeconds:            w.config.DelaySeconds,
			MessageAttributes:       newMessageAttributeMap(msg),
			MessageDeduplicationId:  nil,
			MessageGroupId:          nil,
			MessageSystemAttributes: nil,
		}
		if isQueueFIFO {
			entry.MessageDeduplicationId = msgID
			entry.MessageGroupId = aws.String(msg.StreamKey)
		}
		batchBuf[i] = entry
	}

	out, err := w.client.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		Entries:  batchBuf,
		QueueUrl: aws.String(queueURL),
	})
	if err != nil {
		return err
	} else if len(out.Failed) > 0 {
		errs := &multierror.Error{}
		for _, fail := range out.Failed {
			errs = multierror.Append(errs, errors.New(*fail.Message))
		}
		return errs.ErrorOrNil()
	}
	return nil
}

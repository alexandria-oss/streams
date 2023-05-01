package sns

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/amazon"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/hashicorp/go-multierror"
	jsoniter "github.com/json-iterator/go"
)

// Writer is the Amazon Simple Notification Service (SNS) streams.Writer implementation.
type Writer struct {
	amazon.Writer
	config  amazon.Config
	client  *sns.Client
	baseARN string
}

var _ streams.Writer = Writer{}

// NewWriter allocates an Amazon Simple Notification Service (SNS) concrete implementation of streams.Writer.
func NewWriter(cfg amazon.Config, client *sns.Client) Writer {
	w := Writer{
		config:  cfg,
		client:  client,
		baseARN: fmt.Sprintf("arn:aws:sns:%s:%s", cfg.Region, cfg.AccountID),
	}
	w.WriteFunc = w.write
	return w
}

func (w Writer) write(ctx context.Context, stream string, msgBatch []streams.Message) error {
	isTopicFIFO := strings.HasSuffix(stream, ".fifo")
	batchBuf := make([]types.PublishBatchRequestEntry, len(msgBatch))
	for i, msg := range msgBatch {
		msgJSON, err := jsoniter.Marshal(message{
			Default: string(msg.Data),
		})
		if err != nil {
			return err
		}
		msgID := aws.String(msg.ID)
		msgKey := aws.String(msg.StreamKey)
		entry := types.PublishBatchRequestEntry{
			Id:                     msgID,
			Message:                aws.String(string(msgJSON)),
			MessageAttributes:      newMessageAttributeMap(msg),
			MessageDeduplicationId: nil,
			MessageGroupId:         nil,
			MessageStructure:       aws.String("json"),
			Subject:                msgKey,
		}
		if isTopicFIFO {
			entry.MessageDeduplicationId = msgID
			entry.MessageGroupId = msgKey
		}

		batchBuf[i] = entry
	}

	topicARN := newTopic(w.baseARN, stream)
	out, err := w.client.PublishBatch(ctx, &sns.PublishBatchInput{
		PublishBatchRequestEntries: batchBuf,
		TopicArn:                   aws.String(topicARN),
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

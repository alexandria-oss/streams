package sqs

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/amazon"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// ReaderConfig is the Amazon SQS reader configuration schema.
type ReaderConfig struct {
	amazon.Config
	Logger              *log.Logger   // Logging instance preferably with log level at <<info>>.
	ErrorLogger         *log.Logger   // Logging instance preferably with log level at <<error>>.
	MaxMessagesPerPoll  int32         // Maximum number of message for each polling process (up to 10).
	InFlightInterval    time.Duration // Total time a polling worker will lock a polled message batch, making this batch unavailable to other parallel workers.
	PollParkingDuration time.Duration // Maximum duration for a polling worker to wait for messages.
	PollInterval        time.Duration // Total time the polling worker scheduler will wait between each polling process.
	HandlerTimeout      time.Duration // Maximum duration for message handler processes (streams.ReaderHandleFunc).
}

// Reader is the Amazon Simple Queue Service (SQS) streams.Reader implementation.
type Reader struct {
	config       ReaderConfig
	client       *sqs.Client
	baseQueueURL string

	inFlightWorkers *sync.WaitGroup
}

var _ streams.Reader = Reader{}

// NewReader allocates an Amazon Simple Queue Service (SQS) concrete implementation of streams.Reader.
func NewReader(cfg ReaderConfig, awsCfg aws.Config, client *sqs.Client) Reader {
	if cfg.Logger == nil || cfg.ErrorLogger == nil {
		logger := log.New(os.Stdout, "streams.amazon.sqs: ", 0)
		if cfg.Logger == nil {
			cfg.Logger = logger
		}
		if cfg.ErrorLogger == nil {
			cfg.ErrorLogger = logger
		}
	}
	if cfg.HandlerTimeout == 0 {
		cfg.HandlerTimeout = time.Second * 30
	}
	return Reader{
		config:          cfg,
		client:          client,
		baseQueueURL:    newBaseQueueURL(cfg.Config, awsCfg),
		inFlightWorkers: &sync.WaitGroup{},
	}
}

func (r Reader) Read(ctx context.Context, task streams.ReadTask) (err error) {
mainLoop:
	for {
		select {
		case <-ctx.Done():
			r.config.Logger.Printf("stopping queue polling process")
			break mainLoop
		default:
		}

		r.inFlightWorkers.Add(1)
		queueURL := newQueueURL(r.baseQueueURL, task.Stream)
		out, errRec := r.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:                aws.String(queueURL),
			AttributeNames:          nil,
			MaxNumberOfMessages:     r.config.MaxMessagesPerPoll,
			MessageAttributeNames:   []string{"All"},
			ReceiveRequestAttemptId: nil,
			VisibilityTimeout:       int32(r.config.InFlightInterval.Seconds()),
			WaitTimeSeconds:         int32(r.config.PollParkingDuration.Seconds()),
		})
		if errors.Is(errRec, context.DeadlineExceeded) || errors.Is(errRec, context.Canceled) {
			r.config.Logger.Printf("stopping queue polling process")
			r.inFlightWorkers.Done()
			break
		} else if errRec != nil {
			r.config.ErrorLogger.Print(errRec)
			err = errRec
			time.Sleep(r.config.PollInterval)
			r.inFlightWorkers.Done()
			continue
		} else if len(out.Messages) == 0 {
			time.Sleep(r.config.PollInterval)
			r.inFlightWorkers.Done()
			continue
		}

		go r.schedTask(queueURL, out.Messages, task)
		time.Sleep(r.config.PollInterval)
	}
	r.inFlightWorkers.Wait()
	return
}

type acknowledgeMessage struct {
	messageID string
	receipt   string
}

func (r Reader) schedTask(queueURL string, messages []types.Message, task streams.ReadTask) {
	defer r.inFlightWorkers.Done()
	mu := sync.Mutex{}
	ackBuffer := make([]acknowledgeMessage, 0, len(messages))
	wg := sync.WaitGroup{}
	wg.Add(len(messages))

	for _, msgCp := range messages {
		go func(msg types.Message) {
			defer wg.Done()
			scopedCtx, cancel := context.WithTimeout(context.Background(), r.config.HandlerTimeout)
			defer cancel()
			errHandle := task.Handler(scopedCtx, unmarshalMessage(msg))
			if errHandle != nil {
				// do nothing as developers are able to wrap message handler with middleware functions.
				//
				// This will avoid acknowledging the message and thus, message handler will get retried
				// once VisibilityTimeout (InFlightTimeout) ends.
				return
			}
			mu.Lock()
			ackBuffer = append(ackBuffer, acknowledgeMessage{
				messageID: *msg.MessageId,
				receipt:   *msg.ReceiptHandle,
			})
			mu.Unlock()

		}(msgCp)
	}
	wg.Wait()

	batchEntries := make([]types.DeleteMessageBatchRequestEntry, len(ackBuffer))
	for i, ack := range ackBuffer {
		batchEntries[i] = types.DeleteMessageBatchRequestEntry{
			Id:            aws.String(ack.messageID),
			ReceiptHandle: aws.String(ack.receipt),
		}
	}

	scopedCtx, cancel := context.WithTimeout(context.Background(), r.config.HandlerTimeout)
	defer cancel()
	out, err := r.client.DeleteMessageBatch(scopedCtx, &sqs.DeleteMessageBatchInput{
		Entries:  batchEntries,
		QueueUrl: aws.String(queueURL),
	})
	if err != nil {
		r.config.ErrorLogger.Print(err)
		return
	} else if len(out.Failed) > 0 {
		r.config.ErrorLogger.Printf("failed to acknowledge <%d> messages", len(out.Failed))
	}
	r.config.Logger.Printf("acknowledged <%d>, not acknowledged <%d>", len(out.Successful), len(out.Failed))
}

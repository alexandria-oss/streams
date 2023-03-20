package kafka

import (
	"context"

	"github.com/alexandria-oss/streams"
	"github.com/segmentio/kafka-go"
)

type Writer struct {
	kWriter *kafka.Writer
}

var _ streams.Writer = Writer{}

func NewWriter(kafkaWriter *kafka.Writer) Writer {
	return Writer{kWriter: kafkaWriter}
}

func (w Writer) Write(ctx context.Context, msgBatch []streams.Message) error {
	return w.kWriter.WriteMessages(ctx, marshalMessageBatch(msgBatch)...)
}

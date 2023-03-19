package persistence

import (
	"github.com/alexandria-oss/streams"
)

// NewTransportMessage allocates a TransportMessage from a streams.Message.
func NewTransportMessage(msg streams.Message) *TransportMessage {
	return &TransportMessage{
		MessageId:   msg.ID,
		StreamName:  msg.StreamName,
		ContentType: msg.ContentType,
		Data:        msg.Data,
	}
}

// NewTransportMessageBatch allocates a TransportMessageBatch for each of streams.Message(s).
func NewTransportMessageBatch(msgs []streams.Message) *TransportMessageBatch {
	if len(msgs) == 0 {
		return nil
	}

	buf := make([]*TransportMessage, 0, len(msgs))
	for _, msg := range msgs {
		buf = append(buf, NewTransportMessage(msg))
	}
	return &TransportMessageBatch{
		Messages: buf,
	}
}

// NewMessage allocates a streams.Message from a TransportMessage.
func NewMessage(batchMsg *TransportMessage) streams.Message {
	return streams.Message{
		ID:          batchMsg.GetMessageId(),
		StreamName:  batchMsg.GetStreamName(),
		ContentType: batchMsg.GetContentType(),
		Data:        batchMsg.GetData(),
	}
}

// NewMessages allocates a streams.Message slice from TransportMessageBatch.Messages.
func NewMessages(batch *TransportMessageBatch) []streams.Message {
	msgs := batch.GetMessages()
	if len(msgs) == 0 {
		return nil
	}

	buf := make([]streams.Message, 0, len(msgs))
	for _, msg := range msgs {
		buf = append(buf, NewMessage(msg))
	}

	return buf
}

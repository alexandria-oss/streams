package persistence

import (
	"github.com/alexandria-oss/streams"
)

func NewTransportMessage(msg streams.Message) *TransportMessage {
	return &TransportMessage{
		MessageId:   msg.ID,
		StreamName:  msg.StreamName,
		ContentType: msg.ContentType,
		Data:        msg.Data,
	}
}

func NewTransportMessages(msgs []streams.Message) *TransportMessageBatch {
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

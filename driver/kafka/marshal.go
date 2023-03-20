package kafka

import (
	"strconv"

	"github.com/alexandria-oss/streams"
	"github.com/segmentio/kafka-go"
)

const (
	fixedHeaderCount             = 2
	fixedHeaderReaderInjectCount = 6
	messageIDHeaderKey           = "message_id"
	contentTypeHeaderKey         = "content_type"
)

func marshalMessageHeaders(msg streams.Message) []kafka.Header {
	headers := make([]kafka.Header, 0, fixedHeaderCount+len(msg.Headers))
	headers = append(headers,
		kafka.Header{
			Key:   messageIDHeaderKey,
			Value: []byte(msg.ID),
		},
		kafka.Header{
			Key:   contentTypeHeaderKey,
			Value: []byte(msg.ContentType),
		},
	)
	for k, v := range msg.Headers {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}
	return headers
}

func marshalMessageBatch(msgs []streams.Message) []kafka.Message {
	buf := make([]kafka.Message, 0, len(msgs))
	for _, msg := range msgs {
		buf = append(buf, kafka.Message{
			Topic:   msg.StreamName,
			Key:     []byte(msg.StreamKey),
			Value:   msg.Data,
			Headers: marshalMessageHeaders(msg),
			Time:    msg.Time,
		})
	}
	return buf
}

func unmarshalMessage(msg kafka.Message) streams.Message {
	var (
		messageID   string
		contentType string
	)
	headers := make(map[string]string, (len(msg.Headers)+fixedHeaderReaderInjectCount)-fixedHeaderCount<<0)
	headers[HeaderPartitionID] = strconv.Itoa(msg.Partition)
	headers[HeaderCurrentOffset] = strconv.Itoa(int(msg.Offset))
	headers[HeaderHighWaterMarkOffset] = strconv.Itoa(int(msg.HighWaterMark))
	for _, h := range msg.Headers {
		val := string(h.Value)
		switch h.Key {
		case messageIDHeaderKey:
			messageID = val
		case contentTypeHeaderKey:
			contentType = val
		default:
			headers[h.Key] = val
		}
	}

	return streams.Message{
		ID:          messageID,
		StreamName:  msg.Topic,
		StreamKey:   string(msg.Key),
		Headers:     headers,
		ContentType: contentType,
		Data:        msg.Value,
		Time:        msg.Time,
	}
}

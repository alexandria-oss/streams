package sqs

import (
	"strconv"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/amazon"
	"github.com/alexandria-oss/streams/internal/genericutil"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func newMessageAttributeMap(msg streams.Message) map[string]types.MessageAttributeValue {
	buf := make(map[string]types.MessageAttributeValue, len(msg.Headers)+5)
	defaultDataType := aws.String("String")
	buf[amazon.HeaderMessageID] = types.MessageAttributeValue{
		StringValue: aws.String(msg.ID),
		DataType:    defaultDataType,
	}
	buf[amazon.HeaderStreamName] = types.MessageAttributeValue{
		StringValue: aws.String(msg.StreamName),
		DataType:    defaultDataType,
	}
	buf[amazon.HeaderStreamKey] = types.MessageAttributeValue{
		StringValue: aws.String(msg.StreamKey),
		DataType:    defaultDataType,
	}
	buf[amazon.HeaderContentType] = types.MessageAttributeValue{
		StringValue: aws.String(msg.ContentType),
		DataType:    defaultDataType,
	}
	buf[amazon.HeaderMessageTime] = types.MessageAttributeValue{
		StringValue: aws.String(strconv.FormatInt(msg.Time.UnixMilli(), 10)),
		DataType:    defaultDataType,
	}
	for k, v := range msg.Headers {
		buf[k] = types.MessageAttributeValue{
			StringValue: aws.String(v),
			DataType:    defaultDataType,
		}
	}
	return buf
}

func appendMessageHeaders(rawHeaders map[string]types.MessageAttributeValue, msg *streams.Message) {
	for key, rawHead := range rawHeaders {
		switch key {
		case amazon.HeaderMessageID:
			msg.ID = genericutil.SafeDerefPtr(rawHead.StringValue)
		case amazon.HeaderStreamName:
			msg.StreamName = genericutil.SafeDerefPtr(rawHead.StringValue)
		case amazon.HeaderStreamKey:
			msg.StreamKey = genericutil.SafeDerefPtr(rawHead.StringValue)
		case amazon.HeaderContentType:
			msg.ContentType = genericutil.SafeDerefPtr(rawHead.StringValue)
		case amazon.HeaderMessageTime:
			timeMilli, _ := strconv.ParseInt(genericutil.SafeDerefPtr(rawHead.StringValue), 10, 64)
			msg.Time = time.UnixMilli(timeMilli)
		default:
			msg.Headers[key] = genericutil.SafeDerefPtr(rawHead.StringValue)
		}
	}
}

func unmarshalMessage(rawMsg types.Message) streams.Message {
	// 6 as types.Message has 4 fields to be appended into headers
	headers := make(map[string]string, len(rawMsg.MessageAttributes)+6)
	headers[HeaderMessageID] = genericutil.SafeDerefPtr(rawMsg.MessageId)
	headers[HeaderMessageAttributesMD5] = genericutil.SafeDerefPtr(rawMsg.MD5OfMessageAttributes)
	headers[HeaderMessageReceiptHandle] = genericutil.SafeDerefPtr(rawMsg.ReceiptHandle)
	headers[HeaderMessageBodyMD5] = genericutil.SafeDerefPtr(rawMsg.MD5OfBody)
	msg := streams.Message{
		Data:    []byte(genericutil.SafeDerefPtr(rawMsg.Body)),
		Headers: headers,
	}
	appendMessageHeaders(rawMsg.MessageAttributes, &msg)
	return msg
}

package sns

import (
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/amazon"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

// The SNS transport message schema.
//
// Reference docs: https://docs.aws.amazon.com/sns/latest/api/API_Publish.html#API_Publish_Example_2
type message struct {
	Default   string `json:"default"`
	Email     string `json:"email"`
	EmailJSON string `json:"email-json"`
	HTTP      string `json:"http"`
	HTTPS     string `json:"https"`
	SQS       string `json:"sqs"`
}

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
		StringValue: aws.String(msg.Time.Format(time.RFC3339)),
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

//go:build integration

package sqs_test

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/amazon"
	streamsqs "github.com/alexandria-oss/streams/driver/amazon/sqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type readerSuite struct {
	suite.Suite
	client        *sqs.Client
	awsCfg        aws.Config
	inFlightTests sync.WaitGroup

	accountID string
	stream    string
	queueURL  string
}

func TestReader(t *testing.T) {
	suite.Run(t, &readerSuite{})
}

func (s *readerSuite) SetupSuite() {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("fake", "fake", "")),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               "http://localhost:4566",
					HostnameImmutable: false,
					PartitionID:       "aws",
					SigningName:       "",
					SigningRegion:     "us-east-1",
					SigningMethod:     "",
					Source:            0,
				}, nil
			})),
	)
	s.Require().NoError(err)
	s.awsCfg = cfg
	s.client = sqs.NewFromConfig(cfg)
	s.accountID = "000000000000"
	s.stream = "alexandria-queue-read"
	s.createTopic()
	s.seedMessages()
}

func (s *readerSuite) createTopic() {
	out, err := s.client.CreateQueue(context.Background(), &sqs.CreateQueueInput{
		QueueName:  aws.String(s.stream),
		Attributes: nil,
		Tags:       nil,
	})
	s.Require().NoError(err)
	s.queueURL = *out.QueueUrl
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

func (s *readerSuite) seedMessages() {
	msgs := []streams.Message{
		{
			ID:          "123",
			StreamName:  s.stream,
			StreamKey:   "joe_doe",
			Headers:     nil,
			ContentType: "application/text",
			Data:        []byte("this is joe doe"),
			Time:        time.Time{},
			DecodedData: nil,
		},
		{
			ID:          "456",
			StreamName:  s.stream,
			StreamKey:   "jane_doe",
			Headers:     nil,
			ContentType: "application/text",
			Data:        []byte("this is jane doe"),
			Time:        time.Time{},
			DecodedData: nil,
		},
	}

	out, err := s.client.SendMessageBatch(context.Background(), &sqs.SendMessageBatchInput{
		Entries: []types.SendMessageBatchRequestEntry{
			{
				Id:                      &msgs[0].ID,
				MessageBody:             aws.String(string(msgs[0].Data)),
				DelaySeconds:            0,
				MessageAttributes:       newMessageAttributeMap(msgs[0]),
				MessageSystemAttributes: nil,
			},
			{
				Id:                      &msgs[1].ID,
				MessageBody:             aws.String(string(msgs[1].Data)),
				DelaySeconds:            0,
				MessageAttributes:       newMessageAttributeMap(msgs[1]),
				MessageSystemAttributes: nil,
			},
		},
		QueueUrl: aws.String(s.stream),
	})
	s.Require().NoError(err)
	if len(out.Failed) > 0 {
		s.T().Error(*out.Failed[0].Message)
	}
	s.Require().Len(out.Successful, 2)
}

func (s *readerSuite) TearDownSuite() {
	s.inFlightTests.Wait()
	_, err := s.client.DeleteQueue(context.Background(),
		&sqs.DeleteQueueInput{QueueUrl: aws.String(s.queueURL)})
	s.Assert().NoError(err)
}

func (s *readerSuite) SetupTest() {
	s.inFlightTests.Add(1)
}

func (s *readerSuite) TearDownTest() {

}

func (s *readerSuite) TestRead() {
	reader := streamsqs.NewReader(streamsqs.ReaderConfig{
		Config: amazon.Config{
			AccountID: s.accountID,
			Region:    "us-east-1",
		},
		Logger:              nil,
		ErrorLogger:         nil,
		MaxMessagesPerPoll:  10,
		InFlightInterval:    0,
		PollParkingDuration: 0,
		PollInterval:        0,
		HandlerTimeout:      0,
	}, s.awsCfg, s.client)

	rootCtx, cancelCtx := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelCtx()
	wg := sync.WaitGroup{}
	wg.Add(2)
	var procMessages atomic.Uint64
	task := streams.ReadTask{
		Stream: s.stream,
		Handler: func(ctx context.Context, msg streams.Message) error {
			defer procMessages.Add(1)
			defer wg.Done()
			switch msg.StreamKey {
			case "joe_doe":
				assert.Equal(s.T(), "this is joe doe", string(msg.Data))
				assert.Equal(s.T(), "123", msg.ID)
				assert.Equal(s.T(), "application/text", msg.ContentType)
				assert.Equal(s.T(), "joe_doe", msg.StreamKey)
				assert.Equal(s.T(), s.stream, msg.StreamName)
				assert.NotEmpty(s.T(), msg.Headers[streamsqs.HeaderMessageID])
				assert.NotEmpty(s.T(), msg.Headers[streamsqs.HeaderMessageReceiptHandle])
				assert.NotEmpty(s.T(), msg.Headers[streamsqs.HeaderMessageAttributesMD5])
				assert.NotEmpty(s.T(), msg.Headers[streamsqs.HeaderMessageBodyMD5])
			case "jane_doe":
				assert.Equal(s.T(), "this is jane doe", string(msg.Data))
				assert.Equal(s.T(), "456", msg.ID)
				assert.Equal(s.T(), "application/text", msg.ContentType)
				assert.Equal(s.T(), "jane_doe", msg.StreamKey)
				assert.Equal(s.T(), s.stream, msg.StreamName)
				assert.NotEmpty(s.T(), msg.Headers[streamsqs.HeaderMessageID])
				assert.NotEmpty(s.T(), msg.Headers[streamsqs.HeaderMessageReceiptHandle])
				assert.NotEmpty(s.T(), msg.Headers[streamsqs.HeaderMessageAttributesMD5])
				assert.NotEmpty(s.T(), msg.Headers[streamsqs.HeaderMessageBodyMD5])
			default:
				s.T().Error("no existing stream key received")
			}
			return nil
		},
		ExternalArgs: map[streams.ArgKey]any{},
	}

	go func() {
		err := reader.Read(rootCtx, task)
		if procMessages.Load() < 2 {
			// avoid deadlocks
			wg.Done()
			wg.Done()
			s.T().Error("no messages were received")
		}
		assert.NoError(s.T(), err)
		s.inFlightTests.Done()
	}()
	wg.Wait()
	s.T().Log("at end")
}

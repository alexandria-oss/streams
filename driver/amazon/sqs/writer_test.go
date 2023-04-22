//go:build integration

package sqs_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/amazon"
	streamsqs "github.com/alexandria-oss/streams/driver/amazon/sqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/suite"
)

type writerSuite struct {
	suite.Suite
	client        *sqs.Client
	awsCfg        aws.Config
	inFlightTests sync.WaitGroup

	accountID string
	stream    string
	queueURL  string
}

func TestWriter(t *testing.T) {
	suite.Run(t, &writerSuite{})
}

func (s *writerSuite) SetupSuite() {
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
	s.stream = "alexandria-queue-write"
	s.createTopic()
}

func (s *writerSuite) createTopic() {
	out, err := s.client.CreateQueue(context.Background(), &sqs.CreateQueueInput{
		QueueName:  aws.String(s.stream),
		Attributes: nil,
		Tags:       nil,
	})
	s.Require().NoError(err)
	s.queueURL = *out.QueueUrl
}

func (s *writerSuite) TearDownSuite() {
	s.inFlightTests.Wait()
	_, err := s.client.DeleteQueue(context.Background(),
		&sqs.DeleteQueueInput{QueueUrl: aws.String(s.queueURL)})
	s.Assert().NoError(err)
}

func (s *writerSuite) SetupTest() {
	s.inFlightTests.Add(1)
}

func (s *writerSuite) TearDownTest() {
	s.inFlightTests.Done()
}

func (s *writerSuite) TestWrite() {
	w := streamsqs.NewWriter(streamsqs.WriterConfig{
		Config: amazon.Config{
			AccountID: s.accountID,
			Region:    "us-east-1",
		},
		DelaySeconds: 0,
	}, s.awsCfg, s.client)
	err := w.Write(context.TODO(), []streams.Message{
		{
			ID:         "123",
			StreamName: s.stream,
			StreamKey:  "test_route_key",
			Headers: map[string]string{
				"test_header":   "foo",
				"test_header_2": "bar",
			},
			ContentType: "application/json",
			Data:        []byte("{\"message\":\"foo example\"}"),
			Time:        time.Time{},
			DecodedData: nil,
		},
	})
	s.Assert().NoError(err)
}

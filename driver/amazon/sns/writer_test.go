//go:build integration

package sns_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/amazon"
	streamsns "github.com/alexandria-oss/streams/driver/amazon/sns"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/suite"
)

type writerSuite struct {
	suite.Suite
	client        *sns.Client
	inFlightTests sync.WaitGroup

	accountID string
	stream    string
	topicName string
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
			})))
	s.Require().NoError(err)
	s.client = sns.NewFromConfig(cfg)
	s.accountID = "000000000000"
	s.stream = "alexandria-topic"
	s.createTopic()
}

func (s *writerSuite) createTopic() {
	out, err := s.client.CreateTopic(context.Background(), &sns.CreateTopicInput{
		Name:                 aws.String(s.stream),
		Attributes:           nil,
		DataProtectionPolicy: nil,
		Tags:                 nil,
	})
	s.Require().NoError(err)
	s.topicName = *out.TopicArn
}

func (s *writerSuite) TearDownSuite() {
	s.inFlightTests.Wait()
	_, err := s.client.DeleteTopic(context.Background(),
		&sns.DeleteTopicInput{TopicArn: aws.String(s.topicName)})
	s.Require().NoError(err)
}

func (s *writerSuite) SetupTest() {
	s.inFlightTests.Add(1)
}

func (s *writerSuite) TearDownTest() {
	s.inFlightTests.Done()
}

func (s *writerSuite) TestWrite() {
	w := streamsns.NewWriter(amazon.Config{
		AccountID: s.accountID,
		Region:    "us-east-1",
	}, s.client)
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

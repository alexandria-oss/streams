//go:build integration

package dynamodb_test

import (
	"context"
	"strings"
	"testing"

	streamsdynamo "github.com/alexandria-oss/streams/driver/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/suite"
)

type dedupeStorageSuit struct {
	suite.Suite
	client    *dynamodb.Client
	tableName string
}

func TestDeduplicationStorage(t *testing.T) {
	suite.Run(t, &dedupeStorageSuit{})
}

func (s *dedupeStorageSuit) SetupSuite() {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("fake", "fake", "TOKEN")),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               "http://localhost:8001",
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
	s.client = dynamodb.NewFromConfig(cfg)
	s.tableName = "deduplication-storage"
	s.runMigrations()
}

func (s *dedupeStorageSuit) runMigrations() {
	_, err := s.client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("message_id"),
				AttributeType: "S",
			},
			{
				AttributeName: aws.String("worker_id"),
				AttributeType: "S",
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("message_id"),
				KeyType:       "HASH",
			},
			{
				AttributeName: aws.String("worker_id"),
				KeyType:       "RANGE",
			},
		},
		TableName:              aws.String(s.tableName),
		BillingMode:            "",
		GlobalSecondaryIndexes: nil,
		LocalSecondaryIndexes:  nil,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		SSESpecification:    nil,
		StreamSpecification: nil,
		TableClass:          "",
		Tags:                nil,
	})
	if err != nil && !strings.Contains(err.Error(), "ResourceInUseException") {
		s.Fail(err.Error())
	}
}

func (s *dedupeStorageSuit) TearDownSuite() {
	_, err := s.client.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
		TableName: aws.String(s.tableName),
	})
	s.Assert().NoError(err)
}

func (s *dedupeStorageSuit) TestStorage() {
	dedupeStorage := streamsdynamo.NewDeduplicationStorage(streamsdynamo.DeduplicationStorageConfig{
		TableName:   s.tableName,
		Logger:      nil,
		ErrorLogger: nil,
	}, s.client)
	worker := "worker-0"
	messageID := "123"
	isDupe, err := dedupeStorage.IsDuplicated(context.TODO(), worker, messageID)
	s.Assert().Nil(err)
	s.Assert().False(isDupe)

	dedupeStorage.Commit(context.TODO(), worker, messageID)
	isDupe, err = dedupeStorage.IsDuplicated(context.TODO(), worker, messageID)
	s.Assert().NoError(err)
	s.Assert().True(isDupe)
}

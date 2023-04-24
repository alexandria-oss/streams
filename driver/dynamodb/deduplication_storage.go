package dynamodb

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/alexandria-oss/streams"
)

// DeduplicationStorageConfig is the configuration schema for Amazon DynamoDB streams.DeduplicationStorage implementation.
type DeduplicationStorageConfig struct {
	TableName   string
	Logger      *log.Logger
	ErrorLogger *log.Logger
}

// DeduplicationStorage is the Amazon DynamoDB streams.DeduplicationStorage
type DeduplicationStorage struct {
	client   *dynamodb.Client
	cfg      DeduplicationStorageConfig
	tableRef *string
}

var _ streams.DeduplicationStorage = DeduplicationStorage{}

func NewDeduplicationStorage(cfg DeduplicationStorageConfig, client *dynamodb.Client) DeduplicationStorage {
	if cfg.Logger == nil || cfg.ErrorLogger == nil {
		logger := log.New(os.Stdout, "streams.dynamodb: ", 0)
		if cfg.Logger == nil {
			cfg.Logger = logger
		}
		if cfg.ErrorLogger == nil {
			cfg.ErrorLogger = logger
		}
	}
	return DeduplicationStorage{
		client:   client,
		cfg:      cfg,
		tableRef: aws.String(cfg.TableName),
	}
}

func (d DeduplicationStorage) Commit(ctx context.Context, workerID, messageID string) {
	_, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		Item: map[string]types.AttributeValue{
			"message_id": &types.AttributeValueMemberS{
				Value: messageID,
			},
			"worker_id": &types.AttributeValueMemberS{
				Value: workerID,
			},
		},
		TableName:                   d.tableRef,
		ConditionExpression:         nil,
		ConditionalOperator:         "",
		Expected:                    nil,
		ExpressionAttributeNames:    nil,
		ExpressionAttributeValues:   nil,
		ReturnConsumedCapacity:      "",
		ReturnItemCollectionMetrics: "",
		ReturnValues:                "",
	})
	if err != nil {
		d.cfg.ErrorLogger.Printf("failed to commit message, error %s", err.Error())
		return
	}

	d.cfg.Logger.Printf("committed message with id <%s> and worker id <%s>", workerID, messageID)
}

func (d DeduplicationStorage) IsDuplicated(ctx context.Context, workerID, messageID string) (bool, error) {
	out, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"message_id": &types.AttributeValueMemberS{
				Value: messageID,
			},
			"worker_id": &types.AttributeValueMemberS{
				Value: workerID,
			},
		},
		TableName:                d.tableRef,
		AttributesToGet:          nil,
		ConsistentRead:           nil,
		ExpressionAttributeNames: nil,
		ProjectionExpression:     nil,
		ReturnConsumedCapacity:   "",
	})

	if err != nil {
		d.cfg.ErrorLogger.Printf("failed to get message commit, error %s", err.Error())
		return false, err
	}

	return len(out.Item) >= 2, nil
}

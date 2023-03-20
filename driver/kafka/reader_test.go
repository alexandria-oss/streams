//go:build integration

package kafka_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alexandria-oss/streams"
	streamskafka "github.com/alexandria-oss/streams/driver/kafka"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type readerSuite struct {
	suite.Suite
	inFlightTests    sync.WaitGroup
	topicPartitioned string
	topicGroup       string
	address          string
}

func TestReader(t *testing.T) {
	suite.Run(t, &readerSuite{})
}

func (s *readerSuite) SetupSuite() {
	s.address = "localhost:9092"
	s.topicPartitioned = "org.alexandria.integration_test.read_suite_partition"
	s.topicGroup = "org.alexandria.integration_test.read_suite_group"
	createTopic(s.T(), s.address, s.topicPartitioned)
	createTopic(s.T(), s.address, s.topicGroup)
}

func (s *readerSuite) TearDownSuite() {
	s.inFlightTests.Wait()
	deleteTopic(s.T(), s.address, s.topicPartitioned)
	deleteTopic(s.T(), s.address, s.topicGroup)
}

func (s *readerSuite) SetupTest() {
	s.inFlightTests.Add(1)
}

func (s *readerSuite) TearDownTest() {
	s.inFlightTests.Done()
}

func (s *readerSuite) publishMessage(topic, msg string) {
	w := &kafka.Writer{
		Addr:  kafka.TCP(s.address),
		Topic: topic,
	}
	defer w.Close()
	err := w.WriteMessages(context.TODO(), kafka.Message{
		Partition:     0,
		Offset:        0,
		HighWaterMark: 0,
		Key:           []byte("reader_key"),
		Value:         []byte(msg),
		Headers: []kafka.Header{
			{
				Key:   "message_id",
				Value: []byte("123456"),
			},
			{
				Key:   "content_type",
				Value: []byte("application/text"),
			},
			{
				Key:   "test_header_reader",
				Value: []byte("foo"),
			},
		},
		WriterData: nil,
		Time:       time.Time{},
	})
	require.NoError(s.T(), err)
}

func (s *readerSuite) TestReader_Partitioned() {
	s.publishMessage(s.topicPartitioned, "the quick brown fox offset 0") // publish two messages to start reading from pos 1 offset instead 0
	s.publishMessage(s.topicPartitioned, "the quick brown fox offset 1")
	reader := streamskafka.NewReader(streamskafka.ReaderConfig{
		ReaderConfig: kafka.ReaderConfig{
			Brokers: []string{s.address},
		},
	})

	rootCtx, cancelCtx := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelCtx()
	wg := sync.WaitGroup{}
	wg.Add(1)
	task := streams.ReadTask{
		Stream: s.topicPartitioned,
		Handler: func(ctx context.Context, msg streams.Message) error {
			assert.Equal(s.T(), "the quick brown fox offset 1", string(msg.Data))
			assert.Equal(s.T(), "123456", msg.ID)
			assert.Equal(s.T(), "application/text", msg.ContentType)
			assert.Equal(s.T(), "reader_key", msg.StreamKey)
			assert.Equal(s.T(), s.topicPartitioned, msg.StreamName)
			assert.Equal(s.T(), "foo", msg.Headers["test_header_reader"])
			assert.Equal(s.T(), "0", msg.Headers[streamskafka.HeaderPartitionID])
			assert.Equal(s.T(), "1", msg.Headers[streamskafka.HeaderInitialOffset])
			assert.Equal(s.T(), "1", msg.Headers[streamskafka.HeaderCurrentOffset])
			assert.Equal(s.T(), "2", msg.Headers[streamskafka.HeaderHighWaterMarkOffset])
			assert.Empty(s.T(), msg.Headers[streamskafka.HeaderGroupID])
			s.T().Log("exiting partitioned reader test handler")
			wg.Done()
			return nil
		},
		ExternalArgs: map[streams.ArgKey]any{
			streamskafka.ReaderTaskPartitionIDKey:   0,
			streamskafka.ReaderTaskInitialOffsetKey: int64(1),
		},
	}
	go func() {
		if err := reader.Read(rootCtx, task); err != nil {
			assert.NoError(s.T(), err)
		}
	}()
	wg.Wait()
}

func (s *readerSuite) TestReader_Group() {
	s.publishMessage(s.topicGroup, "the quick brown fox offset 0")
	reader := streamskafka.NewReader(streamskafka.ReaderConfig{
		ReaderConfig: kafka.ReaderConfig{
			Brokers:     []string{s.address},
			Topic:       s.topicGroup,
			StartOffset: kafka.LastOffset,
			GroupID:     "test_group",
		},
	})

	rootCtx, cancelCtx := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelCtx()
	wg := sync.WaitGroup{}
	wg.Add(1)
	task := streams.ReadTask{
		Stream: s.topicGroup,
		Handler: func(ctx context.Context, msg streams.Message) error {
			assert.Equal(s.T(), "the quick brown fox offset 0", string(msg.Data))
			assert.Equal(s.T(), "123456", msg.ID)
			assert.Equal(s.T(), "application/text", msg.ContentType)
			assert.Equal(s.T(), "reader_key", msg.StreamKey)
			assert.Equal(s.T(), s.topicGroup, msg.StreamName)
			assert.Equal(s.T(), "foo", msg.Headers["test_header_reader"])
			assert.Equal(s.T(), "0", msg.Headers[streamskafka.HeaderPartitionID])
			assert.Equal(s.T(), "-2", msg.Headers[streamskafka.HeaderInitialOffset])
			assert.Equal(s.T(), "0", msg.Headers[streamskafka.HeaderCurrentOffset])
			assert.Equal(s.T(), "1", msg.Headers[streamskafka.HeaderHighWaterMarkOffset])
			assert.Equal(s.T(), "test_group", msg.Headers[streamskafka.HeaderGroupID])
			s.T().Log("exiting group reader test handler")
			wg.Done()
			return nil
		},
		ExternalArgs: map[streams.ArgKey]any{
			streamskafka.ReaderTaskPartitionIDKey:   0,
			streamskafka.ReaderTaskInitialOffsetKey: kafka.FirstOffset,
			streamskafka.ReaderTaskGroupIDKey:       "test_group",
		},
	}
	go func() {
		if err := reader.Read(rootCtx, task); err != nil {
			assert.NoError(s.T(), err)
		}
	}()
	wg.Wait()
}

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

type writerSuite struct {
	suite.Suite
	inFlightTests sync.WaitGroup
	topic         string
	address       string
}

func TestWriter(t *testing.T) {
	suite.Run(t, &writerSuite{})
}

func (s *writerSuite) SetupSuite() {
	s.address = "localhost:9092"
	s.topic = "org.alexandria.integration_test.write_suite"
	createTopic(s.T(), s.address, s.topic)
}

func (s *writerSuite) TearDownSuite() {
	s.inFlightTests.Wait()
	deleteTopic(s.T(), s.address, s.topic)
}

func (s *writerSuite) SetupTest() {
	s.inFlightTests.Add(1)
}

func (s *writerSuite) TearDownTest() {
	s.inFlightTests.Done()
}

func (s *writerSuite) readFromOffset(offset int64, handler func(msg kafka.Message)) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:                []string{s.address},
		GroupID:                "",
		GroupTopics:            nil,
		Topic:                  s.topic,
		Partition:              0,
		Dialer:                 nil,
		QueueCapacity:          0,
		MinBytes:               0,
		MaxBytes:               0,
		MaxWait:                0,
		ReadBatchTimeout:       0,
		ReadLagInterval:        0,
		GroupBalancers:         nil,
		HeartbeatInterval:      0,
		CommitInterval:         0,
		PartitionWatchInterval: 0,
		WatchPartitionChanges:  false,
		SessionTimeout:         0,
		RebalanceTimeout:       0,
		JoinGroupBackoff:       0,
		RetentionTime:          0,
		StartOffset:            offset,
		ReadBackoffMin:         0,
		ReadBackoffMax:         0,
		Logger:                 nil,
		ErrorLogger:            nil,
		IsolationLevel:         0,
		MaxAttempts:            0,
		OffsetOutOfRangeError:  false,
	})
	defer r.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*15)
	defer cancel()
	for {
		msg, err := r.FetchMessage(ctx)
		require.NoError(s.T(), err)
		handler(msg)
		break
	}
}

func (s *writerSuite) TestWriter() {
	kWriter := &kafka.Writer{
		Addr: kafka.TCP(s.address),
	}
	defer kWriter.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	w := streamskafka.NewWriter(kWriter)
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*15)
	defer cancel()
	err := w.Write(ctx, []streams.Message{
		{
			ID:         "123",
			StreamName: s.topic,
			StreamKey:  "test_route_key",
			Headers: map[string]string{
				"test_header":   "foo",
				"test_header_2": "bar",
			},
			ContentType: "application/text",
			Data:        []byte("the quick brown writer test"),
			Time:        time.Time{},
			DecodedData: nil,
		},
	})
	require.NoError(s.T(), err)

	s.readFromOffset(kafka.FirstOffset, func(msg kafka.Message) {
		assert.Equal(s.T(), s.topic, msg.Topic)
		assert.Equal(s.T(), "test_route_key", string(msg.Key))
		assert.Equal(s.T(), "the quick brown writer test", string(msg.Value))
		require.Len(s.T(), msg.Headers, 4)
		assert.Equal(s.T(), "message_id", msg.Headers[0].Key)
		assert.Equal(s.T(), "123", string(msg.Headers[0].Value))
		assert.Equal(s.T(), "content_type", msg.Headers[1].Key)
		assert.Equal(s.T(), "application/text", string(msg.Headers[1].Value))
		assert.Equal(s.T(), "test_header", msg.Headers[2].Key)
		assert.Equal(s.T(), "foo", string(msg.Headers[2].Value))
		assert.Equal(s.T(), "test_header_2", msg.Headers[3].Key)
		assert.Equal(s.T(), "bar", string(msg.Headers[3].Value))
		wg.Done()
	})
	wg.Wait()
}

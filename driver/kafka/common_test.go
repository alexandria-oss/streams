package kafka_test

import (
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

func createTopic(t *testing.T, addr, topic string) {
	conn, err := kafka.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()
	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:              topic,
		NumPartitions:      1,
		ReplicationFactor:  1,
		ReplicaAssignments: nil,
		ConfigEntries:      nil,
	})
	require.NoError(t, err)
}

func deleteTopic(t *testing.T, addr, topic string) {
	conn, err := kafka.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()
	err = conn.DeleteTopics(topic)
	require.NoError(t, err)
}

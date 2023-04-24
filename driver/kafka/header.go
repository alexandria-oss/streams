package kafka

const (
	// HeaderClientID is a header key passed to reader handlers -through streams.Message's Headers field-.
	// This key represents the connection identifier from the underlying Apache Kafka client.
	HeaderClientID = "streams-kafka-client-id"
	// HeaderGroupID is a header key passed to reader handlers -through streams.Message's Headers field-.
	// This key represents the consumer group identifier the reader instance is in.
	HeaderGroupID = "streams-kafka-group-id"
	// HeaderPartitionID is a header key passed to reader handlers -through streams.Message's Headers field-.
	// This key represents the partition identifier the reader instance is reading from.
	HeaderPartitionID = "streams-kafka-partition-id"
	// HeaderInitialOffset is a header key passed to reader handlers -through streams.Message's Headers field-.
	// This key represents the offset a reader instance started to read messages from the topic's partition append log.
	HeaderInitialOffset = "streams-kafka-init-offset"
	// HeaderCurrentOffset is a header key passed to reader handlers -through streams.Message's Headers field-.
	// This key represents the offset of a message a reader instance is currently reading from.
	HeaderCurrentOffset = "streams-kafka-current-offset"
	// HeaderHighWaterMarkOffset is a header key passed to reader handlers -through streams.Message's Headers field-.
	// This key represents the next offset available from a topic partition's append log.
	HeaderHighWaterMarkOffset = "streams-kafka-hwm-offset"
)

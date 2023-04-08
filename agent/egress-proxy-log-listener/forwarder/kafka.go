package forwarder

import (
	"github.com/alexandria-oss/streams"
	agent "github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener"
	"github.com/alexandria-oss/streams/agent/egress-proxy-wal-listener/typeutils"
	streamskafka "github.com/alexandria-oss/streams/driver/kafka"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

var defaultKafkaLogger = agent.DefaultLogger.
	With().Str("forwarder_driver", "kafka").Timestamp().Logger()

func NewKafka() (streams.Writer, func() error) {
	infoLogger := defaultKafkaLogger.Level(zerolog.InfoLevel).With().Logger()
	errLogger := defaultKafkaLogger.Level(zerolog.ErrorLevel).With().Logger()
	kWriter := &kafka.Writer{
		Addr:  kafka.TCP(viper.GetStringSlice("kafka.addresses")...),
		Topic: "",
		Balancer: kafka.Murmur2Balancer{
			Consistent: viper.GetBool("kafka.balancer.consistent"),
		},
		MaxAttempts:            viper.GetInt("kafka.max_attempts"),
		WriteBackoffMin:        viper.GetDuration("kafka.write_backoff_min"),
		WriteBackoffMax:        viper.GetDuration("kafka.write_backoff_max"),
		BatchSize:              viper.GetInt("kafka.batch_size"),
		BatchBytes:             viper.GetInt64("kafka.batch_bytes"),
		BatchTimeout:           viper.GetDuration("kafka.batch_timeout"),
		ReadTimeout:            viper.GetDuration("kafka.read_timeout"),
		WriteTimeout:           viper.GetDuration("kafka.write_timeout"),
		RequiredAcks:           typeutils.Coalesce(kafka.RequiredAcks(viper.GetInt("kafka.required_acks")), kafka.RequireOne),
		Async:                  false,
		Completion:             nil,
		Compression:            kafka.Snappy,
		Logger:                 &infoLogger,
		ErrorLogger:            &errLogger,
		Transport:              nil,
		AllowAutoTopicCreation: viper.GetBool("kafka.allow_topic_auto_creation"),
	}
	return streamskafka.NewWriter(kWriter), kWriter.Close
}

package kafka

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/internal/genericutil"
	"github.com/segmentio/kafka-go"
)

const (
	ReaderTaskGroupIDKey       streams.ArgKey = "kafka-group-id"
	ReaderTaskPartitionIDKey   streams.ArgKey = "kafka-partition-id"
	ReaderTaskInitialOffsetKey streams.ArgKey = "kafka-init-offset"
)

type ReaderConfig struct {
	kafka.ReaderConfig
	HandlerTimeout time.Duration
}

type Reader struct {
	cfg ReaderConfig
}

var _ streams.Reader = &Reader{}

func NewReader(cfg ReaderConfig) Reader {
	if cfg.Logger == nil || cfg.ErrorLogger == nil {
		logger := log.New(os.Stdout, "streams.kafka: ", 0)
		if cfg.Logger == nil {
			cfg.Logger = logger
		}
		if cfg.ErrorLogger == nil {
			cfg.ErrorLogger = logger
		}
	}
	if cfg.HandlerTimeout == 0 {
		cfg.HandlerTimeout = time.Second * 30
	}
	return Reader{
		cfg: cfg,
	}
}

func (r Reader) Read(ctx context.Context, task streams.ReadTask) (err error) {
	r.cfg.Topic = task.Stream
	r.cfg.GroupID = genericutil.SafeCast[string](task.ExternalArgs[ReaderTaskGroupIDKey])
	r.cfg.Partition = genericutil.SafeCast[int](task.ExternalArgs[ReaderTaskPartitionIDKey])
	r.cfg.StartOffset = genericutil.SafeCast[int64](task.ExternalArgs[ReaderTaskInitialOffsetKey])

	kReader := kafka.NewReader(r.cfg.ReaderConfig)
	defer func() {
		if errClosure := kReader.Close(); errClosure != nil {
			r.cfg.ErrorLogger.Printf("error occurred closing reader, %s", err.Error())
		}
	}()
	if r.cfg.GroupID == "" && r.cfg.StartOffset != 0 {
		// enable partitioned readers to start at a certain offset in Kafka's partition append log
		err = kReader.SetOffset(r.cfg.StartOffset)
		if err != nil {
			return err
		}
	}

	var kMsg kafka.Message
	for {
		kMsg, err = kReader.FetchMessage(ctx)
		if err != nil {
			r.cfg.ErrorLogger.Printf("error occurred while fetching message, %s", err.Error())
			break
		}

		scopedCtx, cancel := context.WithTimeout(ctx, r.cfg.HandlerTimeout)
		msg := unmarshalMessage(kMsg)
		stats := kReader.Stats()
		msg.Headers[HeaderClientID] = stats.ClientID
		msg.Headers[HeaderGroupID] = r.cfg.GroupID
		msg.Headers[HeaderInitialOffset] = strconv.Itoa(int(r.cfg.StartOffset))

		if errHandler := task.Handler(scopedCtx, msg); errHandler != nil {
			cancel()
			return
		}
		cancel()

		var errCommit error
		if r.cfg.GroupID == "" {
			// commit is not available when reading directly from partitions
			errCommit = kReader.SetOffset(kMsg.Offset + 1)
		} else {
			errCommit = kReader.CommitMessages(ctx, kMsg)
		}

		if errCommit != nil {
			r.cfg.ErrorLogger.Printf("error occurred while committing message, %s", errCommit.Error())
		}
	}

	if err != context.Canceled {
		return err
	}

	return nil
}

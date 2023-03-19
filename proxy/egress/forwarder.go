package egress

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/codec"
	"github.com/alexandria-oss/streams/driver/chanbuf"
	"github.com/alexandria-oss/streams/persistence"
	"github.com/eapache/go-resiliency/retrier"
)

const (
	forwarderWorkerStream = "internal.egress.forwarder.scheduler.jobs" // internal forwarder scheduler job queue.
)

// A Forwarder is an internal component used by an egress proxy agent to redirect queued-traffic (i.e. stream messages)
// to a message broker or similar infrastructure.
type Forwarder struct {
	cfg            ForwarderConfig
	workerSchedBus *chanbuf.Bus
	schedReader    streams.Reader
	schedWriter    streams.Writer
}

// A ForwarderConfig is the configuration used by a Forwarder.
type ForwarderConfig struct {
	Storage                   Storage        // A storage a Forwarder instance uses to fetch message batches (queued-traffic).
	Writer                    streams.Writer // A writer implementation a Forwarder instance uses to publish message batches to.
	Codec                     codec.Codec    // Codec used by the egress writer to store traffic messages.
	Logger                    *log.Logger    // Logger to write information to.
	TableName                 string         // Name of the egress table.
	ForwardJobTimeout         time.Duration  // Maximum time duration to wait a forward job to finish.
	ForwardJobTotalRetries    int            // Maximum count a forward job will be retried.
	ForwardJobRetryBackoff    time.Duration  // Initial time duration between each retry process.
	ForwardJobRetryBackoffMax time.Duration  // Maximum time duration between each retry process.
}

func NewForwarderDefaultConfig() ForwarderConfig {
	return ForwarderConfig{
		Storage:                   nil,
		Writer:                    nil,
		Codec:                     codec.ProtocolBuffers{},
		Logger:                    nil,
		TableName:                 DefaultEgressTableName,
		ForwardJobTimeout:         time.Second * 60,
		ForwardJobTotalRetries:    3,
		ForwardJobRetryBackoff:    time.Second * 5,
		ForwardJobRetryBackoffMax: time.Second * 10,
	}
}

func NewForwarder(cfg ForwarderConfig) Forwarder {
	if cfg.Logger == nil {
		cfg.Logger = log.New(os.Stdout, egressForwarderName+": ", 0)
	}

	bus := chanbuf.NewBus(chanbuf.Config{
		ReaderHandlerTimeout: cfg.ForwardJobTimeout,
		Logger:               cfg.Logger,
	})
	return Forwarder{
		cfg:            cfg,
		workerSchedBus: bus,
		schedReader:    chanbuf.NewReader(bus),
		schedWriter:    chanbuf.NewWriter(bus),
	}
}

// Start initializes the Forwarder instance, blocking the I/O.
// The instance contains internal job scheduling mechanisms for asynchronous job processing.
func (e Forwarder) Start() error {
	retry := retrier.New(retrier.LimitedExponentialBackoff(
		e.cfg.ForwardJobTotalRetries, e.cfg.ForwardJobRetryBackoff, e.cfg.ForwardJobRetryBackoffMax),
		retrier.BlacklistClassifier{
			streams.ErrUnrecoverable,
		},
	)
	retry.SetJitter(0.75)
	err := e.schedReader.Read(context.Background(), forwarderWorkerStream,
		streams.WithReaderRetry(retry)(streams.WithReaderErrorLogger(e.cfg.Logger)(e.scheduleJob)))
	if err != nil {
		return err
	}
	e.cfg.Logger.Printf("starting forwarder")
	e.workerSchedBus.Start()
	return nil
}

// Shutdown gracefully shuts down the Forwarder instance.
func (e Forwarder) Shutdown() {
	e.cfg.Logger.Print("shutting forwarder down")
	e.workerSchedBus.Shutdown()
	e.cfg.Logger.Print("forwarder has been terminated")
}

// Forward triggers a new forward job for the specified batch.
func (e Forwarder) Forward(batchID string) error {
	if len(batchID) == 0 {
		return streams.ErrEmptyMessage
	}
	return e.schedWriter.Write(context.Background(), []streams.Message{
		{
			StreamName: forwarderWorkerStream,
			Data:       []byte(batchID),
		},
	})
}

func (e Forwarder) scheduleJob(ctx context.Context, msg streams.Message) error {
	batchID := string(msg.Data)
	return e.sendBatch(ctx, batchID)
}

func (e Forwarder) sendBatch(ctx context.Context, batchID string) (err error) {
	batch, err := e.cfg.Storage.GetBatch(ctx, batchID)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			return
		} else if errCommit := e.cfg.Storage.Commit(ctx, batchID); errCommit != nil {
			err = errCommit
			return
		}

		e.cfg.Logger.Printf("forwarded traffic from batch_id <%s>", batchID)
	}()

	transportBatch := &persistence.TransportMessageBatch{}
	if err = e.cfg.Codec.Decode(batch.TransportBatchRaw, transportBatch); err != nil {
		return streams.ErrUnrecoverableWrap{ParentErr: err}
	}

	return e.cfg.Writer.Write(ctx, persistence.NewMessages(transportBatch))
}

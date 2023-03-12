package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexandria-oss/streams/codec"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/persistence"
)

// A Writer is a SQL database writer.
// This specific kind of streams.Writer is used by systems implementing the transactional outbox messaging pattern.
// More in depth, a Writer instance will attempt to execute a transactional write into an <<egress table>> where all messages
// generated by the system will be stored, so they may be later publish by an <<egress proxy agent>> (aka. log trailing).
//
// Writer instances MUST be used along transaction context functions
// (persistence.SetTransactionContext, persistence.GetTransactionContext). This is because Writer instances obtain
// the sql.Tx instance from the context. If no context is found, then Writer.Write will fail.
//
// Finally, the main reason to apply the transactional outbox pattern is to obtain write atomicity between a database
// and an event bus.
//
// Transactional outbox pattern reference: https://microservices.io/patterns/data/transactional-outbox.html
type Writer struct {
	cfg WriterConfig
}

// A WriterConfig is the Writer configuration.
// Writer uses streams.IdentifierFactory to generate batch identifiers.
type WriterConfig struct {
	streams.WriterConfig
	Codec             codec.Codec // used to encode message batches, so it can be stored on the database (default codec.ProtocolBuffers).
	WriterEgressTable string      // table to write message batches to be later published.
}

func newWriterDefaults() WriterConfig {
	return WriterConfig{
		WriterConfig:      streams.DefaultWriterConfig,
		Codec:             codec.ProtocolBuffers{},
		WriterEgressTable: "streams_egress",
	}
}

// NewWriter allocates a new Writer instance with default configuration but open to apply any WriterOption(s).
func NewWriter(opts ...WriterOption) Writer {
	baseOpts := newWriterDefaults()
	for _, o := range opts {
		o.apply(&baseOpts)
	}
	return Writer{
		cfg: baseOpts,
	}
}

// NewWriterWithConfig allocates a new Writer instance with passed configuration.
func NewWriterWithConfig(cfg WriterConfig) Writer {
	return Writer{
		cfg: cfg,
	}
}

// WithParentConfig sets streams.WriterConfig fields.
//
// This function SHOULD be used (if required) as the following example:
//
// writer := sql.NewWriter().WithParentConfig()
func (w Writer) WithParentConfig(opts ...streams.WriterOption) Writer {
	for _, o := range opts {
		o.ApplyWriterCfg(&w.cfg.WriterConfig)
	}
	return w
}

// Write append a new batch of messages into the egress table.
//
// A transaction context (persistence.SetTransactionContext) MUST be set before calling this routine.
// This is because Writer instances obtain the sql.Tx instance from the context.
// If no context is found, then Writer.Write will fail.
func (w Writer) Write(ctx context.Context, msgBatch []streams.Message) (err error) {
	if len(msgBatch) == 0 {
		return streams.ErrEmptyMessage
	}

	tx, err := persistence.GetTransactionContext[*sql.Tx](ctx)
	if err != nil {
		return err
	}

	batchID, err := w.cfg.IdentifierFactory()
	if err != nil {
		return err
	}

	var msgBatchAny any = msgBatch
	if w.cfg.Codec.ApplicationType() == codec.ProtocolBuffersApplicationType {
		msgBatchAny = persistence.NewTransportMessages(msgBatch)
	}

	encodedData, err := w.cfg.Codec.Encode(msgBatchAny)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO %s(batch_id,message_count,raw_data) VALUES ($1,$2,$3)", w.cfg.WriterEgressTable)
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(batchID, len(msgBatch), encodedData)
	if err != nil {
		return err
	} else if writeRowCount, _ := res.RowsAffected(); writeRowCount <= 0 {
		err = ErrUnableToWriteRows
	}

	return
}

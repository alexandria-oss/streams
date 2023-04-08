package egress

import (
	"context"
	"time"
)

// A Batch is an aggregate of messages written by a system ready to be published to message brokers or
// similar infrastructure.
type Batch struct {
	BatchID           string
	TransportBatchRaw []byte
	InsertTime        time.Time
}

// A Storage is a special kind of storage where traffic is ingested (queued) so an egress proxy agent (or similar artifacts)
// may forward message batches to another set of infrastructure (e.g. message broker).
type Storage interface {
	// GetBatch retrieves specified batch.
	GetBatch(ctx context.Context, batchID string) (Batch, error)
	// Commit Evicts specified batch.
	Commit(ctx context.Context, batchID string) error
}

// A StorageConfig is the main configuration of a Storage.
type StorageConfig struct {
	TableName string
}

type NoopStorage struct {
	WantGetBatch    Batch
	WantGetBatchErr error
	WantCommitErr   error
}

var _ Storage = NoopStorage{}

func (n NoopStorage) GetBatch(_ context.Context, _ string) (Batch, error) {
	return n.WantGetBatch, n.WantGetBatchErr
}

func (n NoopStorage) Commit(_ context.Context, _ string) error {
	return n.WantCommitErr
}

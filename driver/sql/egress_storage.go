package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexandria-oss/streams/proxy/egress"
)

// A EgressStorage is a SQL implementation of proxy.EgressStorage. Enables interaction with
// a stream egress table (aka. outbox).
type EgressStorage struct {
	db  *sql.DB
	cfg egress.StorageConfig
}

var _ egress.Storage = EgressStorage{}

func newEgressStorageDefaults() egress.StorageConfig {
	return egress.StorageConfig{
		TableName: egress.DefaultEgressTableName,
	}
}

// NewEgressStorage allocates a new EgressStorage instance with default configuration but open to apply any proxy.EgressStorageOption(s).
func NewEgressStorage(db *sql.DB, opts ...egress.StorageOption) EgressStorage {
	baseOpts := newEgressStorageDefaults()
	for _, o := range opts {
		o.Apply(&baseOpts)
	}
	return NewEgressStorageWithConfig(db, baseOpts)
}

// NewEgressStorageWithConfig allocates a new EgressStorage instance with a specific proxy.EgressStorageConfig.
func NewEgressStorageWithConfig(db *sql.DB, cfg egress.StorageConfig) EgressStorage {
	return EgressStorage{
		db:  db,
		cfg: cfg,
	}
}

func (e EgressStorage) GetBatch(ctx context.Context, batchID string) (egress.Batch, error) {
	conn, err := e.db.Conn(ctx)
	if err != nil {
		return egress.Batch{}, err
	}
	defer func() {
		if errConn := conn.Close(); errConn != nil {
			err = errConn
		}
	}()

	query := fmt.Sprintf("SELECT batch_id,raw_data FROM %s WHERE batch_id = $1", e.cfg.TableName)
	row := conn.QueryRowContext(ctx, query, batchID)
	if err = row.Err(); err != nil {
		return egress.Batch{}, err
	}

	batch := egress.Batch{}
	return batch, row.Scan(&batch.BatchID, &batch.TransportBatchRaw)
}

func (e EgressStorage) Commit(ctx context.Context, batchID string) error {
	conn, err := e.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if errConn := conn.Close(); errConn != nil {
			err = errConn
		}
	}()

	query := fmt.Sprintf("DELETE FROM %s WHERE batch_id = $1", e.cfg.TableName)
	_, err = conn.ExecContext(ctx, query, batchID)
	return err
}

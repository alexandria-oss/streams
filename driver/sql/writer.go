package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alexandria-oss/streams"
	jsoniter "github.com/json-iterator/go"
)

type Config struct {
	IdentifierFactory streams.IdentifierFactory
	WriterEgressTable string
}

type Writer struct {
	cfg Config
}

func NewWriter(cfg Config) Writer {
	return Writer{
		cfg: cfg,
	}
}

func (w Writer) Write(ctx context.Context, msgBatch []streams.Message) (err error) {
	tx, err := GetTransactionContext[*sql.Tx](ctx)
	if err != nil {
		return err
	}

	batchID, err := w.cfg.IdentifierFactory()
	if err != nil {
		return err
	}

	rawData, err := jsoniter.Marshal(msgBatch)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`INSERT INTO %s(batch_id,raw_data) VALUES ($1,$2)`, w.cfg.WriterEgressTable)
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(batchID, rawData)
	if err != nil {
		return err
	} else if writeRowCount, _ := res.RowsAffected(); writeRowCount <= 0 {
		err = ErrUnableToWriteRows
	}

	return
}

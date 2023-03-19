package sql

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/codec"
	"github.com/alexandria-oss/streams/persistence"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWriter(t *testing.T) {
	writerWithCfg := NewWriterWithConfig(WriterConfig{
		WriterEgressTable: "foo_table",
	})
	assert.Equal(t, "foo_table", writerWithCfg.cfg.WriterEgressTable)

	tests := []struct {
		name         string
		opts         []WriterOption
		validateFunc func(t *testing.T, w Writer)
	}{
		{
			name: "default values",
			opts: nil,
			validateFunc: func(t *testing.T, w Writer) {
				id, _ := w.cfg.IdentifierFactory()
				_, errID := ksuid.Parse(id)
				assert.NoError(t, errID)
				assert.IsType(t, codec.ProtocolBuffers{}, w.cfg.Codec)
				assert.Equal(t, "streams_egress", w.cfg.WriterEgressTable)
			},
		},
		{
			name: "with change",
			opts: []WriterOption{
				WithEgressTable("foo_table"),
				WithCodec(codec.JSON{}),
			},
			validateFunc: func(t *testing.T, w Writer) {
				id, _ := w.cfg.IdentifierFactory()
				_, errID := ksuid.Parse(id)
				assert.NoError(t, errID)
				assert.IsType(t, codec.JSON{}, w.cfg.Codec)
				assert.Equal(t, "foo_table", w.cfg.WriterEgressTable)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWriter(tt.opts...)
			tt.validateFunc(t, w)
		})
	}
}

func TestWriter_Write(t *testing.T) {
	db, mock, errMock := sqlmock.New()
	require.NoError(t, errMock)
	defer db.Close()

	mock.ExpectBegin()

	tx, errTx := db.BeginTx(context.TODO(), &sql.TxOptions{})
	require.NoError(t, errTx)

	mock.ExpectPrepare("INSERT INTO (.+) VALUES (.+)").WillBeClosed().
		ExpectExec().WithArgs("1", 2, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(123, 1))

	tests := []struct {
		name   string
		writer streams.Writer
		ctx    context.Context
		inMsgs []streams.Message
		err    error
	}{
		{
			name:   "no messages",
			writer: NewWriter(),
			ctx:    nil,
			inMsgs: nil,
			err:    streams.ErrEmptyMessage,
		},
		{
			name:   "no tx context",
			writer: NewWriter(),
			ctx:    context.TODO(),
			inMsgs: []streams.Message{
				{
					StreamName:  "",
					ContentType: "",
					Data:        nil,
				},
			},
			err: persistence.ErrTransactionContextNotFound,
		},
		{
			name: "codec fail",
			writer: NewWriter(WithCodec(codec.Mock{
				EncodeError: errors.New("codec error"),
			}), WithIdentifierFactory(func() (string, error) {
				return "1", nil
			}),
			),
			ctx: persistence.SetTransactionContext[*sql.Tx](context.TODO(), tx),
			inMsgs: []streams.Message{
				{
					StreamName:  "bar",
					ContentType: "application/foo",
					Data:        []byte("foo"),
				},
			},
			err: errors.New("codec error"),
		},
		{
			name: "id factory fail",
			writer: NewWriter(WithCodec(codec.Mock{}), WithIdentifierFactory(func() (string, error) {
				return "", errors.New("id factory error")
			})),
			ctx: persistence.SetTransactionContext[*sql.Tx](context.TODO(), tx),
			inMsgs: []streams.Message{
				{
					StreamName:  "bar",
					ContentType: "application/foo",
					Data:        []byte("foo"),
				},
			},
			err: errors.New("id factory error"),
		},
		{
			name: "valid tx",
			writer: NewWriter(WithCodec(codec.Mock{}), WithIdentifierFactory(func() (string, error) {
				return "1", nil
			})),
			ctx: persistence.SetTransactionContext[*sql.Tx](context.TODO(), tx),
			inMsgs: []streams.Message{
				{
					StreamName:  "bar",
					ContentType: "application/foo",
					Data:        []byte("foo"),
				}, {
					StreamName:  "bar",
					ContentType: "application/foo",
					Data:        []byte("foo"),
				},
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.writer.Write(tt.ctx, tt.inMsgs)
			assert.EqualValues(t, tt.err, err)
		})
	}
}

package egress_test

import (
	"testing"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/persistence"
	"github.com/alexandria-oss/streams/proxy/egress"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestEgress(t *testing.T) {
	batchProto := persistence.NewTransportMessageBatch([]streams.Message{
		{
			ID:          "abc",
			StreamName:  "foo",
			ContentType: "application/text",
			Data:        []byte("the quick brown fox"),
		},
	})
	batchBytes, errMarshal := proto.Marshal(batchProto)
	require.NoError(t, errMarshal)

	defaultCfg := egress.NewForwarderDefaultConfig()
	tests := []struct {
		name      string
		inCfg     egress.ForwarderConfig
		inBatchID string
		wantErr   error
	}{
		{
			name: "unrecoverable",
			inCfg: egress.ForwarderConfig{
				Storage: egress.NoopStorage{
					WantGetBatch: egress.Batch{
						BatchID:           "123",
						TransportBatchRaw: batchBytes,
					},
				},
				Writer: streams.NoopWriter{
					WantWriterErr: streams.ErrUnrecoverableWrap{
						ParentErr: streams.ErrEmptyMessage,
					},
				},
				Codec:                  defaultCfg.Codec,
				Logger:                 defaultCfg.Logger,
				TableName:              defaultCfg.TableName,
				ForwardJobTimeout:      defaultCfg.ForwardJobTimeout,
				ForwardJobTotalRetries: 2,
				ForwardJobRetryBackoff: time.Nanosecond,
			},
			inBatchID: "123",
		},
		{
			name: "missing batch id",
			inCfg: egress.ForwarderConfig{
				Storage:                egress.NoopStorage{},
				Writer:                 streams.NoopWriter{},
				Codec:                  defaultCfg.Codec,
				Logger:                 defaultCfg.Logger,
				TableName:              defaultCfg.TableName,
				ForwardJobTimeout:      defaultCfg.ForwardJobTimeout,
				ForwardJobTotalRetries: 2,
				ForwardJobRetryBackoff: time.Nanosecond,
			},
			inBatchID: "",
			wantErr:   streams.ErrEmptyMessage,
		},
		{
			name: "happy path",
			inCfg: egress.ForwarderConfig{
				Storage: egress.NoopStorage{
					WantGetBatch: egress.Batch{
						BatchID:           "123",
						TransportBatchRaw: batchBytes,
					},
				},
				Writer: streams.NoopWriter{
					WantWriterErr: nil,
				},
				Codec:                  defaultCfg.Codec,
				Logger:                 defaultCfg.Logger,
				TableName:              defaultCfg.TableName,
				ForwardJobTimeout:      defaultCfg.ForwardJobTimeout,
				ForwardJobTotalRetries: 1,
				ForwardJobRetryBackoff: time.Nanosecond,
			},
			inBatchID: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fwd := egress.NewForwarder(tt.inCfg)
			go fwd.Start()
			defer fwd.Shutdown()

			err := fwd.Forward(tt.inBatchID)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

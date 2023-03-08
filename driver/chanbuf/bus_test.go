package chanbuf_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/chanbuf"
	"github.com/stretchr/testify/assert"
)

func TestBus(t *testing.T) {
	bus := chanbuf.NewBus(chanbuf.Config{
		QueueBufferFactor:    0,
		ReaderHandlerTimeout: time.Second * 15,
		Logger:               nil,
	})
	bus.Shutdown() // should do nothing
	go bus.Start()
	defer func() {
		bus.Shutdown()
		err := bus.Publish(streams.Message{})
		assert.ErrorIs(t, streams.ErrBusIsShutdown, err)
	}()

	type dummy struct {
		Bar string `json:"bar"`
	}
	dataJSON, _ := json.Marshal(dummy{Bar: "lorem ipsum"})
	type validateReaderFunc func(t *testing.T) streams.ReaderHandleFunc
	tests := []struct {
		name         string
		inStreamName string
		inMsg        streams.Message
		expHandler   validateReaderFunc
		expErr       error
	}{
		{
			name:         "one sub",
			inStreamName: "foo",
			inMsg: streams.Message{
				StreamName: "foo",
				Data:       dataJSON,
			},
			expHandler: func(t *testing.T) streams.ReaderHandleFunc {
				return func(ctx context.Context, msg streams.Message) error {
					assert.Equal(t, "foo", msg.StreamName)
					assert.Equal(t, "{\"bar\":\"lorem ipsum\"}", string(msg.Data))
					return nil
				}
			},
			expErr: nil,
		},
		{
			name:         "two sub",
			inStreamName: "foo",
			inMsg: streams.Message{
				StreamName: "foo",
				Data:       dataJSON,
			},
			expHandler: func(t *testing.T) streams.ReaderHandleFunc {
				return func(ctx context.Context, msg streams.Message) error {
					assert.Equal(t, "foo", msg.StreamName)
					assert.Equal(t, "{\"bar\":\"lorem ipsum\"}", string(msg.Data))
					return nil
				}
			},
			expErr: nil,
		},
		{
			name:         "failing sub",
			inStreamName: "foo",
			inMsg: streams.Message{
				StreamName: "foo",
				Data:       dataJSON,
			},
			expHandler: func(t *testing.T) streams.ReaderHandleFunc {
				return func(ctx context.Context, msg streams.Message) error {
					return errors.New("foo failed")
				}
			},
			expErr: nil,
		},
		{
			name:         "no subscriber",
			inStreamName: "foo",
			inMsg: streams.Message{
				StreamName: "bar",
				Data:       dataJSON,
			},
			expHandler: nil,
			expErr:     nil,
		},
		{
			name:         "no data",
			inStreamName: "foo",
			inMsg: streams.Message{
				StreamName: "bar",
				Data:       nil,
			},
			expHandler: nil,
			expErr:     streams.ErrEmptyMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expHandler != nil {
				bus.Subscribe(tt.inStreamName, tt.expHandler(t))
			}
			assert.ErrorIs(t, tt.expErr, bus.Publish(tt.inMsg))
		})
	}
}

package streams_test

import (
	"context"
	"testing"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/chanbuf"
	"github.com/stretchr/testify/assert"
)

type anyEvent struct {
	ID string `json:"id"`
}

var _ streams.Event = anyEvent{}

func (a anyEvent) GetHeaders() map[string]string {
	return nil
}

func (a anyEvent) GetKey() string {
	return a.ID
}

func TestBus(t *testing.T) {
	reader := chanbuf.NewReader(nil)
	writer := chanbuf.NewWriter(nil)
	go chanbuf.DefaultBus.Start()

	bus := streams.NewBus(writer, reader)
	bus.RegisterEvent(anyEvent{}, "org.alexandria.events")
	bus.Subscribe("any.event", anyEvent{}, func(ctx context.Context, msg streams.Message) error {
		assert.Equal(t, "123-foo", msg.StreamKey)
		return nil
	}).SetArg("some_arg", "test_arg")
	bus.SubscribeTopic("org.alexandria.events", func(ctx context.Context, msg streams.Message) error {
		assert.Equal(t, "123-foo", msg.StreamKey)
		return nil
	})
	go func(bus *streams.Bus) {
		if errStart := bus.Start(); errStart != nil && errStart != streams.ErrNoSubscriberRegistered {
			t.Error(errStart)
			return
		}
	}(&bus)
	defer func(bus *streams.Bus) {
		if errStop := bus.Shutdown(); errStop != nil {
			t.Fatal(errStop)
		}
	}(&bus)
	defer chanbuf.DefaultBus.Shutdown()
	err := bus.Publish(context.TODO(), anyEvent{
		ID: "123-foo",
	})
	if err != nil {
		t.Fatal(err)
	}
}

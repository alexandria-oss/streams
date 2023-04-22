package streams_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alexandria-oss/streams"
)

type fakeEvent struct {
}

var _ streams.Event = fakeEvent{}

func (f fakeEvent) GetHeaders() map[string]string {
	return nil
}

func (f fakeEvent) GetKey() string {
	return ""
}

func TestSubscriberScheduler_Subscribe(t *testing.T) {
	reg := streams.EventRegistry{}
	reg.RegisterEvent(fakeEvent{}, "fake-stream")
	sched := streams.NewSubscriberScheduler(nil, reg)
	out := sched.Subscribe(fakeEvent{}, func(ctx context.Context, msg streams.Message) error {
		return nil
	}).SetArg("test", "this is a test")
	assert.Equal(t, "this is a test", out.ExternalArgs["test"])
}

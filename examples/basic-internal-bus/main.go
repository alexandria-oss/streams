package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/chanbuf"
)

// 1. Define the message schema to be passed through the stream.
//
// This will be later propagated by brokers to readers/consumers/subscribers.
type clickStreamEvent struct {
	ClickedComponent string    `json:"clicked_component"`
	ClickTime        time.Time `json:"click_time"`
}

func main() {
	// 2. Allocate a new Bus instance.
	bus := chanbuf.NewBus(chanbuf.Config{
		QueueBufferFactor: 0, // Set buffered channel allocation capacity. Leave zero value to use unbuffered.
		// Time to wait for each ReaderHandler process. Will send a Done signal to scoped context.Context so
		// internal processes can be shut down gracefully.
		ReaderHandlerTimeout: 0,
		// A logger instance to be used by the Bus. If nil, Bus will use a basic os.Stdout logger.
		Logger: nil,
	})
	// 3. Register the relationships between a stream and its handler(s).
	//
	// A stream MIGHT have up to N handlers (i.e. 1:M, one-to-many).
	//
	// NOTE: Handlers are executed concurrently. Thus, process execution ordering is not guaranteed.
	bus.Subscribe("foo", func(ctx context.Context, msg streams.Message) error {
		// 7. Once a message has been received, the Bus will execute this routine.
		log.Printf("[handler-0] stream: %s | msg: %s", msg.StreamName, string(msg.Data))
		return nil
	})
	bus.Subscribe("foo", func(ctx context.Context, msg streams.Message) error {
		// 7. Once a message has been received, the Bus will execute this routine.
		log.Printf("[handler-1] stream: %s | msg: %s", msg.StreamName, string(msg.Data))
		return nil
	})
	// 4. Define a deferred call to free-up Bus internal resources, avoid leaks (memory/context/goroutine)
	// and handle in-flight reader handler processes gracefully.
	defer bus.Shutdown()

	// 5. Start the Bus instance in a goroutine as Start() starts to consume the internal queue channel, blocking
	// the I/O until Shutdown() is called.
	go bus.Start()

	// 6. Send actual data to the stream.
	data, _ := json.Marshal(clickStreamEvent{
		ClickedComponent: "sign_in",
		ClickTime:        time.Date(2023, 1, 31, 12, 30, 50, 100, time.UTC),
	})
	// NOTE: Publish routine is fire-and-forget, meaning it only cares about writing into the message queue.
	// If any of the subscribers fail, this routine will not handle any of those errors.
	_ = bus.Publish(streams.Message{
		StreamName: "foo",
		Data:       data,
	})
}

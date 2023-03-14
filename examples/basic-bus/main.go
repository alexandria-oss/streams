package main

import (
	"context"
	"log"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/driver/chanbuf"
)

func main() {
	const streamName = "org.alexandria.foo"

	// 1. Allocate a reader to set up default resources and register reader handling functions.
	//
	// A specific chanbuf.Bus can be used instead of a nil value. If a specific bus is used, please
	// use the same instance while declaring new writers as well. Internal mechanisms use the bus instance
	// to communicate.
	var reader streams.Reader = chanbuf.NewReader(nil)
	// 2. Start the bus in a separate routine; chanbuf.Start() is blocking I/O and will be stopped unless
	// chanbuf.Shutdown() is called.
	go chanbuf.Start()
	// 3. Register chanbuf.Shutdown() deferred call to release bus resources and wait for in-flight messages
	// processes to finish (graceful shutdown).
	defer chanbuf.Shutdown()

	// 4. Register reader instances, each reader.Read() call is a new reader task. Each task is handled concurrently
	// and is an independent process. Furthermore, as tasks are executed concurrently, execution ordering is NOT
	// guaranteed.
	_ = reader.Read(context.TODO(), streamName, func(ctx context.Context, msg streams.Message) error {
		log.Printf("[handler-0] at foo stream | %+v", msg)
		log.Printf("[handler-0] at foo stream | %s", string(msg.Data))
		return nil
	})
	_ = reader.Read(context.TODO(), streamName, func(ctx context.Context, msg streams.Message) error {
		log.Printf("[handler-1] at foo stream | %v", msg)
		log.Printf("[handler-1] at foo stream | %s", string(msg.Data))
		return nil
	})

	// 5. Register a writer instance. Each call to writer.Write() will write a new entry into the bus
	// stream. Furthermore, the bus will route this message to any reader listening to the stream (publish/subscribe
	// & fire-and-forget patterns implemented).
	var writer streams.Writer = chanbuf.NewWriter(nil)
	err := writer.Write(context.TODO(), []streams.Message{
		{
			ID:          "123",
			StreamName:  streamName,
			ContentType: "application/text",
			Data:        []byte("the quick brown fox"),
		},
	})
	if err != nil {
		panic(err)
	}

	// 6. Step 3 will be finally called HERE, shutting down the bus instance.
}

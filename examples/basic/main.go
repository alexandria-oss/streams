package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexandria-oss/streams"
	"github.com/alexandria-oss/streams/codec"
	"github.com/alexandria-oss/streams/driver/chanbuf"
)

// 1. Create your event schema, add serialization tags of your like.

type UserCreated struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

// 1.1 Ensure your event schema implements streams.Event interface.
var _ streams.Event = UserCreated{}

// 1.1.2 On this routine, you may append dynamic custom headers to your events.

func (u UserCreated) GetHeaders() map[string]string {
	return nil
}

// 1.1.3 On this routine, set a key that will be used by Writers to route the event to a specific partition.
// This may be unavailable depending on the Writer concrete implementation/configuration.

func (u UserCreated) GetKey() string {
	return u.UserID
}

func main() {
	// 2. Setup both reader and writer instances to be used by the bus.
	var reader streams.Reader = chanbuf.NewReader(nil)
	var writer streams.Writer = chanbuf.NewWriter(nil)

	// 3. Allocate your event bus instance.
	bus := streams.NewBus(writer, reader)

	// 4. Register your subscription(s) to your bus. You may set first argument (topic name) just once as
	// bus.Subscribe will use UserCreated topic once it detects a relation between the schema and
	// the topic, meaning you only need to pass the topic name just one time.
	//
	// Moreover, messages MIGHT come through without a specific order, meaning subscriber 1 could
	// get processed first instead subscriber 0.
	bus.Subscribe("user.created", UserCreated{}, func(ctx context.Context, msg streams.Message) error {
		userEvent := UserCreated{}
		if err := codec.Unmarshal(msg.ContentType, msg.Data, &userEvent); err != nil {
			return err
		}
		log.Printf("[At subscriber 0] %s", userEvent)
		// DO SOMETHING HERE...
		return nil // Returning an error here will NOT acknowledge the message consumption, retrying the process or just failing.
	}).SetArg("group", "send-welcome-email")
	bus.Subscribe("", UserCreated{}, func(ctx context.Context, msg streams.Message) error {
		log.Print("[At subscriber 1]")
		// DO SOMETHING HERE...
		return nil
	}).SetArg("group", "aggregate-user-metrics")

	// 4b. [OPTIONAL] Register all of your event schemas used by readers/writers.
	// bus.RegisterEvent(UserCreated{}, "")

	sysChan := make(chan os.Signal, 2)
	signal.Notify(sysChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		// 5b. This step is required if you used chanbuf.Bus implementation. This
		// routine will start the underlying channel-based bus.
		chanbuf.Start()
	}()
	go func() {
		// 5. Start the bus inside a goroutine as bus MIGHT block I/O.
		if err := bus.Start(); err != nil {
			log.Print(err)
			os.Exit(1)
		}
	}()
	go func() {
		// 6. Publish your messages as expected.
		err := bus.Publish(context.TODO(), UserCreated{
			UserID:      "123",
			DisplayName: "Joe Doe",
		})
		if err != nil {
			log.Print(err)
		}
	}()
	<-sysChan
	// 7. Shutdown bus instances to deallocate its underlying resources.
	chanbuf.Shutdown()
	_ = bus.Shutdown()
}

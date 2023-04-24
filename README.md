# Streams

:envelope: `Streams` is a stream-based communication toolkit made for data-in-motion platforms that enables applications to use streaming communication mechanisms for efficient data transfer.

## Installation

To install the library, you need to have Golang installed in your system. Once you have Golang installed, you can use the following command to install the library:

```bash
go get github.com/alexandria-oss/streams
go get github.com/alexandria-oss/streams/driver/YOUR_DRIVER
```

## Usage

To use library in your application, you need to import it and create a new instance:

```go
import (
    "github.com/alexandria-oss/streams"
    "github.com/alexandria-oss/streams/chanbuf"
)

func main() {
    // Setup both reader and writer instances to be used by the bus. Chanbuf is the in-memory channel-backed driver.
    var reader streams.Reader = chanbuf.NewReader(nil)
    var writer streams.Writer = chanbuf.NewWriter(nil)
    
    bus := streams.NewBus(writer, reader)
}
```

### Sending Data

To send data using the library, you can use the `Send` method:

```go
err := bus.Publish(context.TODO(), YourMessage{})
```

### Receiving Data

To receive data using the library, you need to register a callback function using the `Subscribe` method:

```go
bus.Subscribe("user.created", UserCreated{}, func(ctx context.Context, msg streams.Message) error {
  userEvent := UserCreated{}
  if err := codec.Unmarshal(msg.ContentType, msg.Data, &userEvent); err != nil {
    return err
  }
  log.Printf("[At subscriber 0] %s", userEvent)
  // DO SOMETHING HERE...
  return nil // Returning an error here will NOT acknowledge the message consumption, retrying the process or just failing.
})
```

## Examples

Here are some examples of how you can use the Streaming Communication Library in your application:

### Example 1: Sending and Receiving Data

```go
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

func main() {
    var reader streams.Reader = chanbuf.NewReader(nil)
    var writer streams.Writer = chanbuf.NewWriter(nil)
    
    bus := streams.NewBus(writer, reader)
    
    bus.Subscribe("user.created", UserCreated{}, func(ctx context.Context, msg streams.Message) error {
      userEvent := UserCreated{}
      if err := codec.Unmarshal(msg.ContentType, msg.Data, &userEvent); err != nil {
        return err
      }
      log.Printf("[At subscriber 0] %s", userEvent)
      return nil
    })
  
    sysChan := make(chan os.Signal, 2)
    signal.Notify(sysChan, os.Interrupt, syscall.SIGTERM)
    go func() {
      // This step is required if you used chanbuf.Bus implementation. This
      // routine will start the underlying channel-based bus.
      chanbuf.Start()
    }()
    go func() {
      // Start the bus inside a goroutine as bus MIGHT block I/O.
      if err := bus.Start(); err != nil {
        log.Print(err)
        os.Exit(1)
      }
    }()
    go func() {
      // Publish your messages as expected.
      err := bus.Publish(context.TODO(), UserCreated{
        UserID:      "123",
        DisplayName: "Joe Doe",
      })
      if err != nil {
        log.Print(err)
      }
    }()
    <-sysChan
    // Shutdown bus instances to deallocate its underlying resources.
    chanbuf.Shutdown()
    _ = bus.Shutdown()
}
```

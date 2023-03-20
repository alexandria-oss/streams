# streams

:envelope: Streams is a stream-based communication toolkit made for data-in-motion platforms.

## Architecture Overview

As previously mentioned, `streams` is toolkit library which helps developers to craft event-driven and/or
data-in-motion backed ecosystems. Therefore, `streams` comes with several components available for usage
to deal with specific use cases or overall scenarios.

### Bus

A `bus` is a high-level component which serves as a simplified proxy (_or API_) for both basic and essential 
`streams` operations.

Example of previously mentioned essential operations:

- Publish a message with features such as:
  - Automatic Go-type mapping to stream/topic.
  - Header setup.
  - Schema registry integration.
- Read messages from a stream with features such as:
  - HTTP router/mux-like handler definition.
- Elver

### Writer

A `writer` is a low-level component that writes into streams and may be configured using driver's `WriterOption`(s).

### Reader

A `reader` is a low-level component that reads from streams, depending on the driver implementation, using 
short/long polling or socket-listening mechanisms.
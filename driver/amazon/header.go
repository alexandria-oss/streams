package amazon

const (
	HeaderMessageID   = "streams-message-id"   // The unique identifier of a message.
	HeaderStreamName  = "streams-stream-name"  // Name of the stream of a message.
	HeaderStreamKey   = "streams-stream-key"   // Key of the stream from a message.
	HeaderContentType = "streams-content-type" // Type of data of a content from a message.
	HeaderMessageTime = "streams-message-time" // Timestamp in Unix milliseconds when the message was published.
)

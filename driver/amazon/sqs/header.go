package sqs

const (
	// HeaderMessageID A unique identifier for the message. A MessageId is considered unique across all
	// Amazon Web Services accounts for an extended period of time.
	HeaderMessageID = "sqs-message-id"
	// HeaderMessageAttributesMD5 An MD5 digest of the non-URL-encoded message attribute string. You can use this
	// attribute to verify that Amazon SQS received the message correctly. Amazon SQS
	// URL-decodes the message before creating the MD5 digest. For information about
	// MD5, see RFC1321 (https://www.ietf.org/rfc/rfc1321.txt) .
	HeaderMessageAttributesMD5 = "sqs-message-attributes-md5"
	// HeaderMessageReceiptHandle An identifier associated with the act of receiving the message. A new receipt
	// handle is returned every time you receive a message. When deleting a message,
	// you provide the last received receipt handle to delete the message.
	HeaderMessageReceiptHandle = "sqs-message-receipt-handle"
	// HeaderMessageBodyMD5 An MD5 digest of the non-URL-encoded message body string.
	HeaderMessageBodyMD5 = "sqs-message-body-md5"
)

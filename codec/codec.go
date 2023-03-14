package codec

// A Codec is a device or computer program that encodes or decodes a data stream or signal.
type Codec interface {
	// Encode encodes the given input into an array of bytes using an underlying serialization type.
	Encode(v any) ([]byte, error)
	// Decode decodes the given input of an array of bytes into a Go type using an underlying serialization type.
	//
	// Go type MUST be a pointer reference so serializer can append data into it.
	Decode(src []byte, dst any) error
	// ApplicationType returns the RFC application type of the underlying serializer implementation
	// (e.g. application/json).
	ApplicationType() string
}

package codec

import "errors"

var codecMap = map[string]Codec{
	JSONApplicationType:            JSON{},
	ProtocolBuffersApplicationType: ProtocolBuffers{},
}

// ErrCodecNotFound the specified codec is not registered in streams' codec package.
var ErrCodecNotFound = errors.New("streams: codec not found")

// Unmarshal decodes a message using the codec specified in codecType (e.g. JSONApplicationType).
func Unmarshal[T any](codecType string, src []byte, dst T) error {
	c, ok := codecMap[codecType]
	if !ok {
		return ErrCodecNotFound
	}

	return c.Decode(src, dst)
}

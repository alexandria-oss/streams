package codec

import "google.golang.org/protobuf/proto"

// ProtocolBuffersApplicationType the ProtocolBuffers application type.
//
// Reference: https://github.com/google/protorpc/blob/95849c9a3e414b9ba2d3da91fd850156348fe558/protorpc/protobuf.py#L52
const ProtocolBuffersApplicationType = "application/x-google-protobuf"

// The ProtocolBuffers codec is a free and open-source cross-platform data format used to serialize structured data.
// It is useful in developing programs to communicate with each other over a network or for storing data.
type ProtocolBuffers struct{}

// compile-time assertion
var _ Codec = ProtocolBuffers{}

func newProtobufMessage(v any) (proto.Message, error) {
	if v == nil {
		return nil, ErrInvalidFormat
	}

	msg, ok := v.(proto.Message)
	if !ok {
		return nil, ErrInvalidFormat
	}
	return msg, nil
}

func (b ProtocolBuffers) Encode(v any) ([]byte, error) {
	msg, err := newProtobufMessage(v)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(msg)
}

func (b ProtocolBuffers) Decode(src []byte, dst any) error {
	msg, err := newProtobufMessage(dst)
	if err != nil {
		return err
	}
	return proto.Unmarshal(src, msg)
}

func (b ProtocolBuffers) ApplicationType() string {
	return ProtocolBuffersApplicationType
}

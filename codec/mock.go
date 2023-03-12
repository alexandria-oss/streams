package codec

// A Mock is a dummy type of Codec ready to be used for testing purposes.
type Mock struct {
	EncodeBytes           []byte
	EncodeError           error
	DecodeError           error
	ApplicationTypeString string
}

var _ Codec = Mock{}

func (n Mock) Encode(_ any) ([]byte, error) {
	return n.EncodeBytes, n.EncodeError
}

func (n Mock) Decode(_ []byte, _ any) error {
	return n.DecodeError
}

func (n Mock) ApplicationType() string {
	return n.ApplicationTypeString
}

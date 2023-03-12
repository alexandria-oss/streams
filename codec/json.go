package codec

import jsoniter "github.com/json-iterator/go"

const JSONApplicationType = "application/json"

type JSON struct{}

// compile-time assertion
var _ Codec = JSON{}

func (j JSON) Encode(v any) ([]byte, error) {
	return jsoniter.Marshal(v)
}

func (j JSON) Decode(src []byte, dst any) error {
	return jsoniter.Unmarshal(src, dst)
}

func (j JSON) ApplicationType() string {
	return JSONApplicationType
}

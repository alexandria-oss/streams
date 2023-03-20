package streams

import (
	"testing"

	"github.com/alexandria-oss/streams/codec"
	"github.com/stretchr/testify/assert"
)

func TestWithPublisherOption(t *testing.T) {
	opt := WithIdentifierFactory(NewKSUID)
	cfg := PublisherConfig{}
	opt.apply(&cfg)
	assert.NotNil(t, cfg.IdentifierFactory)

	opt = WithPublisherCodec(codec.JSON{})
	opt.apply(&cfg)
	assert.Equal(t, codec.JSON{}.ApplicationType(), cfg.Codec.ApplicationType())
}

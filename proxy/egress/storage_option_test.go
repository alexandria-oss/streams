package egress_test

import (
	"testing"

	"github.com/alexandria-oss/streams/proxy/egress"
	"github.com/stretchr/testify/assert"
)

func TestWithEgressTable(t *testing.T) {
	a := egress.WithEgressTable("foo")
	cfg := egress.StorageConfig{}
	a.Apply(&cfg)
	assert.Equal(t, "foo", cfg.TableName)
}

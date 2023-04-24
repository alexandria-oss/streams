package streams_test

import (
	"context"
	"testing"
	"time"

	"github.com/alexandria-oss/streams"
	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedDeduplicationStorage(t *testing.T) {
	cache, err := bigcache.New(context.TODO(), bigcache.DefaultConfig(time.Minute*30))
	require.NoError(t, err)
	defer cache.Close()

	var dedupeStorage streams.DeduplicationStorage
	dedupeStorage = streams.NewEmbeddedDeduplicationStorage(streams.EmbeddedDeduplicationStorageConfig{
		Logger:       nil,
		ErrorLogger:  nil,
		KeyDelimiter: "",
	}, cache)

	worker := "worker-0"
	messageID := "123"
	isDupe, err := dedupeStorage.IsDuplicated(context.TODO(), worker, messageID)
	assert.Nil(t, err)
	assert.False(t, isDupe)

	dedupeStorage.Commit(context.TODO(), worker, messageID)

	isDupe, err = dedupeStorage.IsDuplicated(context.TODO(), worker, messageID)
	assert.NoError(t, err)
	assert.True(t, isDupe)
}

package streams

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/allegro/bigcache/v3"
)

// A DeduplicationStorage is a special kind of storage used by a data-in motion system to
// keep track of duplicate messages.
//
// It is recommended to use in-memory (or at least external) storages
// to increase performance significantly and reduce main database backpressure.
type DeduplicationStorage interface {
	// Commit registers and acknowledges a message has been processed correctly.
	Commit(ctx context.Context, workerID, messageID string)
	// IsDuplicated indicates if a message has been processed before.
	IsDuplicated(ctx context.Context, workerID, messageID string) (bool, error)
}

// EmbeddedDeduplicationStorageConfig is the EmbeddedDeduplicationStorage schema configuration.
type EmbeddedDeduplicationStorageConfig struct {
	Logger       *log.Logger
	ErrorLogger  *log.Logger
	KeyDelimiter string
}

// A EmbeddedDeduplicationStorage is the in-memory concrete implementation of DeduplicationStorage using
// package allegro.BigCache as high-performance underlying storage.
//
// Consider by using this storage, your computing instance becomes stateful, meaning if the node gets down,
// the deduplicated message database will be dropped as well.
type EmbeddedDeduplicationStorage struct {
	cache *bigcache.BigCache
	cfg   EmbeddedDeduplicationStorageConfig
}

const embeddedDeduplicationStorageKeyDelimiter = "#"

var _ DeduplicationStorage = EmbeddedDeduplicationStorage{}

// NewEmbeddedDeduplicationStorage allocates an in-memory DeduplicationStorage instance.
func NewEmbeddedDeduplicationStorage(cfg EmbeddedDeduplicationStorageConfig, cache *bigcache.BigCache) EmbeddedDeduplicationStorage {
	if cfg.Logger == nil || cfg.ErrorLogger == nil {
		logger := log.New(os.Stdout, "streams: ", 0)
		if cfg.Logger == nil {
			cfg.Logger = logger
		}
		if cfg.ErrorLogger == nil {
			cfg.ErrorLogger = logger
		}
	}
	if cfg.KeyDelimiter == "" {
		cfg.KeyDelimiter = embeddedDeduplicationStorageKeyDelimiter
	}
	return EmbeddedDeduplicationStorage{
		cache: cache,
		cfg:   cfg,
	}
}

func (e EmbeddedDeduplicationStorage) Commit(_ context.Context, workerID, messageID string) {
	key := strings.Join([]string{messageID, workerID}, e.cfg.KeyDelimiter)
	// using cache as hash set
	if err := e.cache.Set(key, nil); err != nil {
		e.cfg.ErrorLogger.Printf("failed to commit message, error %s", err.Error())
		return
	}

	e.cfg.Logger.Printf("committed message with id <%s> and worker id <%s>", workerID, messageID)
}

func (e EmbeddedDeduplicationStorage) IsDuplicated(_ context.Context, workerID, messageID string) (bool, error) {
	key := strings.Join([]string{messageID, workerID}, e.cfg.KeyDelimiter)
	if _, err := e.cache.Get(key); err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

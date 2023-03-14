package streams

import (
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/ksuid"
)

// An IdentifierFactory is a small component function that generates unique identifiers.
type IdentifierFactory func() (string, error)

var _ IdentifierFactory = NewUUID

// NewUUID generates a universally unique identifier (UUID).
func NewUUID() (string, error) {
	id, err := uuid.NewUUID()
	return id.String(), err
}

var _ IdentifierFactory = NewKSUID

// NewKSUID generates a k-sortable unique identifier (KSUID).
func NewKSUID() (string, error) {
	id, err := ksuid.NewRandomWithTime(time.Now().UTC())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

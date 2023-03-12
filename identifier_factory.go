package streams

import (
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/ksuid"
)

type IdentifierFactory func() (string, error)

var _ IdentifierFactory = NewUUID

func NewUUID() (string, error) {
	id, err := uuid.NewUUID()
	return id.String(), err
}

var _ IdentifierFactory = NewKSUID

func NewKSUID() (string, error) {
	id, err := ksuid.NewRandomWithTime(time.Now().UTC())
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

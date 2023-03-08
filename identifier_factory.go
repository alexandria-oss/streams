package streams

import "github.com/google/uuid"

type IdentifierFactory func() (string, error)

var _ IdentifierFactory = NewUUID

func NewUUID() (string, error) {
	id, err := uuid.NewUUID()
	return id.String(), err
}

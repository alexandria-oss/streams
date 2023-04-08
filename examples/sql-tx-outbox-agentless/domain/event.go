package domain

type Event interface {
	GetKey() string
	GetStreamName() string
}

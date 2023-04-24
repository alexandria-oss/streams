package streams

import (
	"context"
	"time"

	"github.com/alexandria-oss/streams/codec"
)

// PublisherConfig is the Publisher configuration.
type PublisherConfig struct {
	IdentifierFactory IdentifierFactory
	Codec             codec.Codec
}

// A Publisher is a high-level component which writes Event(s) into topics (streams).
// Depending on the underlying Writer, publish routines will write Event(s) in batches or one-by-one.
type Publisher struct {
	writer    Writer
	eventReg  EventRegistry
	idFactory IdentifierFactory
	codec     codec.Codec
}

func newPublisherDefaults() PublisherConfig {
	return PublisherConfig{
		IdentifierFactory: NewUUID,
		Codec:             codec.JSON{},
	}
}

// NewPublisher allocates a new Publisher instance ready to be used. Specify options to customize default
// configurations.
func NewPublisher(w Writer, eventReg EventRegistry, options ...PublisherOption) Publisher {
	cfg := newPublisherDefaults()
	for _, opt := range options {
		opt.apply(&cfg)
	}
	return Publisher{
		writer:    w,
		eventReg:  eventReg,
		idFactory: cfg.IdentifierFactory,
		codec:     cfg.Codec,
	}
}

// builds a Message out from an Event.
func (p Publisher) newMessage(event Event) (Message, error) {
	msgID, err := p.idFactory()
	if err != nil {
		return Message{}, err
	}

	topic, err := p.eventReg.GetEventTopic(event)
	if err != nil {
		return Message{}, err
	}

	encodedMsg, err := p.codec.Encode(event)
	if err != nil {
		return Message{}, err
	}

	return Message{
		ID:          msgID,
		StreamName:  topic,
		StreamKey:   event.GetKey(),
		Headers:     event.GetHeaders(), // might require to merge maps here if we want to use `streams` internal headers.
		ContentType: p.codec.ApplicationType(),
		Data:        encodedMsg,
		Time:        time.Now().UTC(),
	}, nil
}

// builds a Message out from an Event with specified topic.
func (p Publisher) newMessageWithTopic(topic string, event Event) (Message, error) {
	msgID, err := p.idFactory()
	if err != nil {
		return Message{}, err
	}

	encodedMsg, err := p.codec.Encode(event)
	if err != nil {
		return Message{}, err
	}

	return Message{
		ID:          msgID,
		StreamName:  topic,
		StreamKey:   event.GetKey(),
		Headers:     event.GetHeaders(), // might require to merge maps here if we want to use `streams` internal headers.
		ContentType: p.codec.ApplicationType(),
		Data:        encodedMsg,
		Time:        time.Now().UTC(),
	}, nil
}

// Publish writes Event(s) into a topic specified on each Event.
func (p Publisher) Publish(ctx context.Context, events ...Event) error {
	msgBuf := make([]Message, 0, len(events))
	for _, ev := range events {
		msg, err := p.newMessage(ev)
		if err != nil {
			return err
		}
		msgBuf = append(msgBuf, msg)
	}

	return p.writer.Write(ctx, msgBuf)
}

// PublishToTopic writes Event(s) into a specific topic.
func (p Publisher) PublishToTopic(ctx context.Context, topic string, events ...Event) error {
	msgBuf := make([]Message, 0, len(events))
	for _, ev := range events {
		msg, err := p.newMessageWithTopic(topic, ev)
		if err != nil {
			return err
		}
		msgBuf = append(msgBuf, msg)
	}

	return p.writer.Write(ctx, msgBuf)
}

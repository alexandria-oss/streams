package messaging

import (
	"context"
	"sample/domain"

	"github.com/alexandria-oss/streams"
	jsoniter "github.com/json-iterator/go"
)

func WriteEvents(ctx context.Context, w streams.Writer, ag domain.Aggregate) error {
	events := ag.PullEvents()
	batch := make([]streams.Message, 0, len(events))
	for _, event := range events {
		dataJSON, errJSON := jsoniter.Marshal(event)
		if errJSON != nil {
			continue
		}
		batch = append(batch, streams.Message{
			StreamName:  event.GetStreamName(),
			ContentType: "application/json",
			Data:        dataJSON,
		})
	}

	return w.Write(ctx, batch)
}

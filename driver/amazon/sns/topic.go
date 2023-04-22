package sns

import (
	"fmt"
	"strings"
)

var topicReplacer = strings.NewReplacer(".", "-")

func newTopic(baseARN, streamName string) string {
	streamName = topicReplacer.Replace(streamName)
	return fmt.Sprintf("%s:%s", baseARN, streamName)
}

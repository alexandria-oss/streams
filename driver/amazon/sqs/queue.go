package sqs

import (
	"fmt"
	"strings"

	"github.com/alexandria-oss/streams/driver/amazon"
	"github.com/aws/aws-sdk-go-v2/aws"
)

var queueNameReplacer = strings.NewReplacer(".", "-")

func newBaseQueueURL(cfg amazon.Config, awsCfg aws.Config) string {
	baseQueueURL := fmt.Sprintf("https://sqs.%s.amazonaws.com/%s", cfg.Region, cfg.AccountID)
	if awsCfg.EndpointResolverWithOptions != nil {
		ep, err := awsCfg.EndpointResolverWithOptions.ResolveEndpoint("sqs", awsCfg.Region)
		if err == nil {
			baseQueueURL = ep.URL + "/" + cfg.AccountID
		}
	}
	return baseQueueURL
}

func newQueueURL(baseURL, streamName string) string {
	streamName = queueNameReplacer.Replace(streamName)
	return fmt.Sprintf("%s/%s", baseURL, streamName)
}

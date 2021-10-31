package sns

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/innolight/goevents"
)

func NewQueueProvider(options ...Option) goevents.QueueProvider {
	conf := conf{}
	for _, apply := range options {
		apply(&conf)
	}
	if len(conf.AwsConfigs) == 0 {
		conf.AwsConfigs = append(conf.AwsConfigs, &aws.Config{Region: aws.String("eu-central-1")})
	}

	return &queueProvider{conf: conf}
}

type queueProvider struct {
	conf conf
}

func (f queueProvider) Get(ctx context.Context, topicARN string) (goevents.Queue, error) {
	awsSession, err := session.NewSession(f.conf.AwsConfigs...)
	if err != nil {
		return nil, err
	}
	svc := sns.New(awsSession)

	// Validate the ARN is valid
	_, err = svc.GetTopicAttributesWithContext(ctx, &sns.GetTopicAttributesInput{
		TopicArn: &topicARN,
	})
	if err != nil {
		return nil, fmt.Errorf("SNS topic ARN validation failed: %w", err)
	}

	return &topic{
		topicARN: topicARN,
		sns:      svc,
	}, nil
}

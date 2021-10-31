package sqs

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/innolight/goevents"
)

func NewQueueProvider(options ...Option) goevents.QueueProvider {
	conf := conf{
		ReceiveMaxNumberOfMessage: 10,
		ReceiveWaitTimeSeconds:    20,
	}
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

func (f queueProvider) Get(ctx context.Context, name string) (goevents.Queue, error) {
	awsSession, err := session.NewSession(f.conf.AwsConfigs...)
	if err != nil {
		return nil, err
	}
	svc := sqs.New(awsSession)
	queueURl, err := getQueueUrl(svc, name)
	if err != nil {
		return nil, err
	}
	return &queue{
		queueURL:            queueURl,
		sqs:                 svc,
		maxNumberOfMessages: f.conf.ReceiveMaxNumberOfMessage,
		waitTimeSeconds:     f.conf.ReceiveWaitTimeSeconds,
	}, nil
}

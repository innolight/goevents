package sqs

import (
	"context"
	"github.com/innolight/goevents"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type conf struct {
	AwsConfigs                []*aws.Config
	ReceiveMaxNumberOfMessage int
	ReceiveWaitTimeSeconds    int
}

type Option func(conf *conf)

func WithAwsConfig(awsConf *aws.Config) Option {
	return func(conf *conf) {
		conf.AwsConfigs = append(conf.AwsConfigs, awsConf)
	}
}

func WithReceiveMaxNumberOfMessage(i int) Option {
	return func(conf *conf) {
		conf.ReceiveMaxNumberOfMessage = i
	}
}

func WithReceiveWaitTimeSeconds(i int) Option {
	return func(conf *conf) {
		conf.ReceiveWaitTimeSeconds = i
	}
}

func NewTopicFactory(options ...Option) goevents.TopicFactory {
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

	return &queueFactory{conf: conf}
}

type queueFactory struct {
	conf conf
}

func (f queueFactory) Get(ctx context.Context, name string) (goevents.Topic, error) {
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

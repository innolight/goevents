package sqs

import "github.com/aws/aws-sdk-go/aws"

type Option func(conf *conf)

type conf struct {
	AwsConfigs                []*aws.Config
	ReceiveMaxNumberOfMessage int
	ReceiveWaitTimeSeconds    int
	MappingSNSMessageEnabled  bool
}

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

// If the events from the SQS queue are sourced from SNS, setting this option will automatically
// extract the SNS message as event's body.
func WithMappingSNSMessageEnabled() Option {
	return func(conf *conf) {
		conf.MappingSNSMessageEnabled = true
	}
}

package sns

import "github.com/aws/aws-sdk-go/aws"

type Option func(conf *conf)

type conf struct {
	AwsConfigs []*aws.Config
}

func WithAwsConfig(awsConf *aws.Config) Option {
	return func(conf *conf) {
		conf.AwsConfigs = append(conf.AwsConfigs, awsConf)
	}
}

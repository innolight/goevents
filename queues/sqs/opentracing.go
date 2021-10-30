package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type messageAttributes map[string]*sqs.MessageAttributeValue

func (s messageAttributes) Set(key, val string) {
	if len(val) == 0 {
		return
	}
	s[key] = &sqs.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(val),
	}
}

func (s messageAttributes) ForeachKey(handler func(key, val string) error) error {
	for k, value := range s {
		var v string
		if value.StringValue != nil {
			v = *value.StringValue
		}
		if err := handler(k, v); err != nil {
			return err
		}
	}
	return nil
}

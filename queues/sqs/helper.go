package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

func getQueueUrl(svc sqsiface.SQSAPI, queueName string) (string, error) {
	resultURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", fmt.Errorf("cannot get SQS queue %s: %v", queueName, err)
	}
	if resultURL == nil {
		return "", fmt.Errorf("cannot get SQS queue %s: resultURL is nil", queueName)
	}
	if resultURL.QueueUrl == nil {
		return "", fmt.Errorf("cannot get SQS queue %s: QueueUrl is nil", queueName)
	}
	return *resultURL.QueueUrl, nil
}

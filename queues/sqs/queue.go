package sqs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/innolight/goevents"
	"log"
)

type queue struct {
	queueURL            string
	sqs                 sqsiface.SQSAPI
	maxNumberOfMessages int
	waitTimeSeconds     int
}

func (q *queue) Send(ctx context.Context, event goevents.Event) (err error) {
	bytes, err := event.Body()
	if err != nil {
		return err
	}
	_, err = q.sqs.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(string(bytes)),
		QueueUrl:    aws.String(q.queueURL),
	})
	return err
}

func (q *queue) Receive(ctx context.Context) ([]goevents.EventEnvelop, error) {
	sqsMessages, err := q.sqs.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q.queueURL),
		MaxNumberOfMessages: aws.Int64(int64(q.maxNumberOfMessages)),
		WaitTimeSeconds:     aws.Int64(int64(q.waitTimeSeconds)),
		AttributeNames:      []*string{},
		// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ReceiveMessage.html
		MessageAttributeNames: aws.StringSlice([]string{"All"}),
	})
	if err != nil {
		return nil, fmt.Errorf("sqs.ReceiveMessage :%v", err)
	}

	var results []goevents.EventEnvelop

	for idx := range sqsMessages.Messages {
		// explicitly create a copy
		m := sqsMessages.Messages[idx]

		resultChannel := make(chan error)
		go func() {
			err := <-resultChannel
			close(resultChannel)
			if err != nil {
				return
			}
			// TODO: add logging and retry for acknowledge messages
			_, err = q.sqs.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(q.queueURL),
				ReceiptHandle: m.ReceiptHandle,
			})
			log.Printf("Acknowledged processed message: %s\n", *m.Body)
		}()
		results = append(results, goevents.EventEnvelop{
			Result: resultChannel,
			Event:  goevents.StringEvent(ctx, *m.Body),
		})
	}

	return results, nil
}

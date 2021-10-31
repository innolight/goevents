package sns

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/innolight/goevents"
)

type topic struct {
	topicARN string
	sns      snsiface.SNSAPI
}

func (q *topic) Send(ctx context.Context, event goevents.Event) (err error) {
	bytes, err := event.Body()
	if err != nil {
		return err
	}
	_, err = q.sns.PublishWithContext(ctx, &sns.PublishInput{
		Message:  aws.String(string(bytes)),
		TopicArn: aws.String(q.topicARN),
	})
	return err
}

func (q *topic) Receive(ctx context.Context) ([]goevents.EventEnvelop, error) {
	return nil, fmt.Errorf("receive from SNS is not supported")
}

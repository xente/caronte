package metricstores

import (
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
)

var sqsSession sync.Once
var targetSQS *sqs.SQS
var queueUrl *string

type SQSStore interface {
	Query(specs MetricSpecs) (float64, error)
}

type MetricSQSStore struct {
	QueueName string
	QueueUrl  string
}

func (p MetricSQSStore) Query(specs MetricSpecs) (float64, error) {

	sqsSession.Do(func() {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		targetSQS = sqs.New(sess)

		result, _ := targetSQS.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: aws.String(specs.SQSStore.QueueName),
		})
		queueUrl = result.QueueUrl
	})

	attributes := sqs.GetQueueAttributesInput{
		QueueUrl: queueUrl,
		AttributeNames: []*string{
			aws.String(specs.Query),
		},
	}

	totalValue := 0.0
	//Fix needed to ensure that the sqs results does not return 0 when the queue contains messages
	for i := 0; i < 3; i++ {
		resp, err := targetSQS.GetQueueAttributes(&attributes)
		if err != nil {
			zap.S().Debug(err)
		}

		value, _ := strconv.ParseFloat(*resp.Attributes[specs.Query], 64)
		if totalValue < value {
			totalValue = value
		}

	}

	return totalValue, nil
}

package metrics_publisher

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	sqsApproximateNumberOfMessages = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "caronte_sqs_ApproximateNumberOfMessages",
		Help: "Total SQS messages",
	})
)
var (
	sqsApproximateNumberOfMessagesDelayed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "caronte_sqs_ApproximateNumberOfMessagesDelayed",
		Help: "Total SQS messages",
	})
)
var (
	sqsReceiveMessageWaitTimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "caronte_sqs_ReceiveMessageWaitTimeSeconds",
		Help: "Total SQS messages",
	})
)
var (
	sqsApproximateNumberOfMessagesNotVisible = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "caronte_sqs_ApproximateNumberOfMessagesNotVisible",
		Help: "Total SQS messages",
	})
)

func SQSrecordMetrics(queueMame string, sleepTime int) {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	targetSQS := sqs.New(sess)

	result, _ := targetSQS.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueMame),
	})

	attributes := sqs.GetQueueAttributesInput{
		QueueUrl: result.QueueUrl,
		AttributeNames: []*string{
			aws.String("All"),
		},
	}

	for {
		resp, err := targetSQS.GetQueueAttributes(&attributes)
		if err != nil {
			zap.S().Debug(err)
		} else {

			value, _ := strconv.ParseFloat(*resp.Attributes["ApproximateNumberOfMessages"], 64)
			sqsApproximateNumberOfMessages.Set(value)
			value, _ = strconv.ParseFloat(*resp.Attributes["ApproximateNumberOfMessagesDelayed"], 64)
			sqsApproximateNumberOfMessagesDelayed.Set(value)
			value, _ = strconv.ParseFloat(*resp.Attributes["ReceiveMessageWaitTimeSeconds"], 64)
			sqsReceiveMessageWaitTimeSeconds.Set(value)
			value, _ = strconv.ParseFloat(*resp.Attributes["ApproximateNumberOfMessagesNotVisible"], 64)
			sqsApproximateNumberOfMessagesNotVisible.Set(value)
		}
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

}
